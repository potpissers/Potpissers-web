package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {

}

func Handler(w http.ResponseWriter, r *http.Request) {
	defer postgresPool.Close()
	println("init done")

	const mojangUsernameProxyEndpoint = "/api/proxy/mojang/username/"
	http.HandleFunc(mojangUsernameProxyEndpoint, func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(minecraftUsernameLookupUrl + strings.TrimPrefix(r.URL.Path, mojangUsernameProxyEndpoint))
		if err != nil {
			log.Println(err)
			return
		} else {
			defer resp.Body.Close()

			// TODO -> headers ?
			w.WriteHeader(resp.StatusCode)
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	http.HandleFunc("/api/donations/payments", func(w http.ResponseWriter, r *http.Request) {
		type donateRequest struct {
			Username       string `json:"username"`
			LineItemName   string `json:"line_item_name"`
			LineItemAmount int    `json:"line_item_amount"`

			lineItemCostInCents int
		}
		var attemptedDonationRequests []donateRequest
		err := json.NewDecoder(r.Body).Decode(&attemptedDonationRequests)
		if err != nil {
			log.Println(err)
			return
		}
		var successfulDonationRequests []donateRequest
		var mutex sync.Mutex
		var waitGroup sync.WaitGroup
		for _, request := range attemptedDonationRequests {
			var potentialFinalRequest donateRequest
			for _, data := range lineItemDatas {
				if data.GameModeName+"-"+data.ItemName == request.LineItemName {
					potentialFinalRequest.LineItemName = request.LineItemName
					if data.IsPlural {
						potentialFinalRequest.LineItemAmount = int(math.Max(float64(request.LineItemAmount), 1))
					} else {
						potentialFinalRequest.LineItemAmount = 1 // TODO -> allow 0 amount
					}
					potentialFinalRequest.lineItemCostInCents = data.ItemPriceInCents

					waitGroup.Add(1)
					go func() {
						defer waitGroup.Done()
						for {
							resp, err := getMojangApiUuidRequest(request.Username)
							if err != nil {
								log.Println(err)
							} else {
								statusCode := resp.StatusCode
								err = resp.Body.Close()
								if err != nil {
									log.Println(err)
								}

								if statusCode == 200 {
									potentialFinalRequest.Username = request.Username
									mutex.Lock()
									successfulDonationRequests = append(successfulDonationRequests, potentialFinalRequest)
									mutex.Unlock() // TODO -> this loses the correct order
									break
								} else if statusCode == 404 {
									break
								}
							}
							time.Sleep(time.Second * 2)
						}
					}()
				}
			}
		}
		waitGroup.Wait()
		if len(successfulDonationRequests) == 0 {
			log.Println("err: invalid donationRequest line item")
			return // door nigga from game of thrones
		} else {
			type LineItem struct {
				Quantity       string `json:"quantity"`
				ItemType       string `json:"item_type"`
				Name           string `json:"name"`
				BasePriceMoney money  `json:"base_price_money"`
			}
			var lineItems []LineItem
			for _, lineItem := range successfulDonationRequests {
				lineItems = append(lineItems, LineItem{
					Quantity: strconv.Itoa(lineItem.LineItemAmount),
					ItemType: "ITEM",
					Name:     lineItem.LineItemName + "," + lineItem.Username,
					BasePriceMoney: money{
						Amount:   lineItem.lineItemCostInCents,
						Currency: "USD",
					},
				})
			}
			type Order struct {
				LocationID string     `json:"location_id"`
				LineItems  []LineItem `json:"line_items"`
			}
			type AcceptedPaymentMethods struct {
				AfterpayClearpay bool `json:"afterpay_clearpay"`
				ApplePay         bool `json:"apple_pay"`
				CashAppPay       bool `json:"cash_app_pay"`
				GooglePay        bool `json:"google_pay"`
			}
			type CheckoutOptions struct {
				AllowTipping           bool                   `json:"allow_tipping"`
				AcceptedPaymentMethods AcceptedPaymentMethods `json:"accepted_payment_methods"`
				AskForShippingAddress  bool                   `json:"ask_for_shipping_address"`
				EnableCoupon           bool                   `json:"enable_coupon"`
				EnableLoyalty          bool                   `json:"enable_loyalty"`
				MerchantSupportEmail   string                 `json:"merchant_support_email"`
				RedirectURL            string                 `json:"redirect_url"`
			}
			reqData := struct {
				CheckoutOptions CheckoutOptions `json:"checkout_options"`
				Description     string          `json:"description"`
				Order           Order           `json:"order"`
			}{
				CheckoutOptions: CheckoutOptions{
					AllowTipping: true,
					AcceptedPaymentMethods: AcceptedPaymentMethods{
						AfterpayClearpay: false,
						ApplePay:         true,
						CashAppPay:       true,
						GooglePay:        true,
					},
					AskForShippingAddress: false,
					EnableCoupon:          false,
					EnableLoyalty:         false,
					MerchantSupportEmail:  "potpissers@gmail.com",
					RedirectURL:           "potpissers.com/donations",
				},
				Description: "hey",
				Order: Order{
					LocationID: os.Getenv("SQUARE_LOCATION_ID"),
					LineItems:  lineItems,
				},
			}
			reqBody, err := json.Marshal(reqData)
			if err != nil {
				log.Println(err)
				return
			}
			req, err := http.NewRequest("POST", "https://connect.squareup.com/v2/online-checkout/payment-links", bytes.NewBuffer(reqBody))
			if err != nil {
				log.Println(err)
				return
			}

			req.Header.Set("Square-Version", "2025-02-20")
			req.Header.Set("Authorization", "Bearer "+os.Getenv("SQUARE_ACCESS_TOKEN"))
			req.Header.Set("Content-Type", "application/json")

			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()

			var paymentLinkResp struct {
				PaymentLink struct {
					ID              string          `json:"id"`
					Version         int             `json:"version"`
					Description     string          `json:"description"`
					OrderID         string          `json:"order_id"`
					CheckoutOptions CheckoutOptions `json:"checkout_options"`
					URL             string          `json:"url"`
					LongURL         string          `json:"long_url"`
					CreatedAt       time.Time       `json:"created_at"`
				} `json:"payment_link"`
				RelatedResources struct {
					Orders []struct {
						LocationID string `json:"location_id"`
						LineItems  []struct {
							Quantity       string `json:"quantity"`
							ItemType       string `json:"item_type"`
							Name           string `json:"name"`
							BasePriceMoney money  `json:"base_price_money"`

							UID                      string `json:"uid"`
							VariationTotalPriceMoney money  `json:"variation_total_price_money"`
							GrossSalesMoney          money  `json:"gross_sales_money"`
							TotalTaxMoney            money  `json:"total_tax_money"`
							TotalDiscountMoney       money  `json:"total_discount_money"`
							TotalMoney               money  `json:"total_money"`
							TotalServiceChargeMoney  money  `json:"total_service_charge_money"`
						} `json:"line_items"`
						ID     string `json:"id"`
						Source struct {
							Name string `json:"name"`
						} `json:"source"`
						Fulfillments []struct {
							UID   string `json:"uid"`
							Type  string `json:"type"`
							State string `json:"state"`
						} `json:"fulfillments"`
						NetAmounts struct {
							TotalMoney         money `json:"total_money"`
							TaxMoney           money `json:"tax_money"`
							DiscountMoney      money `json:"discount_money"`
							TipMoney           money `json:"tip_money"`
							ServiceChargeMoney money `json:"service_charge_money"`
						} `json:"net_amounts"`
						CreatedAt               time.Time `json:"created_at"`
						UpdatedAt               time.Time `json:"updated_at"`
						State                   string    `json:"state"`
						Version                 int       `json:"version"`
						TotalMoney              money     `json:"total_money"`
						TotalTaxMoney           money     `json:"total_tax_money"`
						TotalDiscountMoney      money     `json:"total_discount_money"`
						TotalTipMoney           money     `json:"total_tip_money"`
						TotalServiceChargeMoney money     `json:"total_service_charge_money"`
						NetAmountDueMoney       money     `json:"net_amount_due_money"`
					} `json:"orders"`
				} `json:"related_resources"`
			}
			err = json.NewDecoder(resp.Body).Decode(&paymentLinkResp)
			if err != nil {
				log.Println(err)
				return
			}
			_, err = w.Write([]byte(paymentLinkResp.PaymentLink.URL))
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	println("main api done")

	home = getMainTemplateBytes("hub")
	println("home template done")
	hcf = getMainTemplateBytes("hcf" + currentHcfServerName)
	println("hcf template done")
	mz = getMainTemplateBytes("mz")
	println("mz template done")

	for endpoint, bytes := range map[string]*[]byte{
		"/":    &home,
		"/hub": &home,
		"/mz":  &mz,
		"/hcf": &hcf,
	} {
		pointer := bytes
		http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(*pointer)
			handleFatalErr(err)

			handleRedditPostDataUpdate()
			handleDiscordMessagesUpdate(discordGeneralChan, discordGeneralChannelId, &mostRecentDiscordGeneralMessageId, &discordMessages, "general")
			//			handleDiscordMessagesUpdate(discordChangelogChan, discordChangelogChannelId, &mostRecentDiscordChangelogMessageId, &changelog, "changelog")
			handleDiscordMessagesUpdate(discordAnnouncementsChan, discordAnnouncementsChannelId, &mostRecentDiscordAnnouncementsMessageId, &announcements, "announcements")
		})
		println(endpoint + " done")
	}

	http.HandleFunc("/github", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/potpissers", http.StatusMovedPermanently)
	})
	http.HandleFunc("/reddit", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.reddit.com/r/potpissers/", http.StatusMovedPermanently)
	})
	http.HandleFunc("/discord", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://discord.gg/Cqnvktf7EF", http.StatusFound)
	})

	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/static-donate.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/static-init.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/favicon.png", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/mz-map/", http.StripPrefix("/mz-map", http.FileServer(http.Dir(frontendDirName+"/mz-map"))))

	println("starting server")
	log.Println(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem", "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil))
}

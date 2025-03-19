package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const minecraftUsernameLookupUrl = "https://api.minecraftservices.com/minecraft/profile/lookup/name/"

var potpissersTips []string
var cubecoreTips []string
var cubecoreClassTips []string
var mzTips []string

func init() {
	getRowsBlocking("SELECT * FROM get_tips()", func(rows pgx.Rows) {
		var tipMessage struct {
			gameModeName string
			tipTitle     string
			tipMessage   string
		}
		handleFatalPgx(pgx.ForEachRow(rows, []any{&tipMessage.gameModeName, &tipMessage.tipTitle, &tipMessage.tipMessage}, func() error {
			switch tipMessage.gameModeName {
			case "potpissers":
				potpissersTips = append(potpissersTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			case "cubecore":
				cubecoreTips = append(cubecoreTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			case "cubecore_classes":
				cubecoreClassTips = append(cubecoreClassTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			case "mz":
				mzTips = append(mzTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			}
			return nil
		}))
	})
}

type newPlayer struct {
	PlayerUuid string    `json:"playerUuid"`
	Referrer   *string   `json:"referrer"`
	Timestamp  time.Time `json:"timestamp"`
	RowNumber  int       `json:"rowNumber"`
}

var newPlayers = func() []newPlayer {
	var newPlayers []newPlayer
	getRowsBlocking("SELECT * FROM get_12_newest_players()", func(rows pgx.Rows) {
		var death newPlayer
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.PlayerUuid, &death.Referrer, &death.Timestamp, &death.RowNumber}, func() error {
			newPlayers = append(newPlayers, death)
			return nil
		}))
	})
	return newPlayers
}()

func init() {
	http.HandleFunc("/api/players/new", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostJsonPatch[newPlayer](r, func(newT *newPlayer, r *http.Request) error { return json.NewDecoder(r.Body).Decode(&newT) }, &newPlayersMu, &newPlayers)
		home = getHome()
		mz = getMz()
		hcf = getMz()
		// TODO -> sse
	})
}

var newPlayersMu sync.RWMutex

type death struct {
	ServerName        string    `json:"serverName"`
	VictimUserFightId *int      `json:"victimUserFightId"`
	Timestamp         time.Time `json:"timestamp"`
	VictimUuid        string    `json:"victimUuid"`
	// TODO victim inventory
	DeathWorldName string  `json:"deathWorldName"`
	DeathX         int     `json:"deathX"`
	DeathY         int     `json:"deathY"`
	DeathZ         int     `json:"deathZ"`
	DeathMessage   string  `json:"deathMessage"`
	KillerUuid     *string `json:"killerUuid"`
	// TODO killer weapon
	// TODO killer inventory
}

var deathsMu sync.RWMutex
var deaths = func() []death {
	var deaths []death
	getRowsBlocking("SELECT * FROM get_12_latest_network_deaths()", func(rows pgx.Rows) {
		var death death
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
			deaths = append(deaths, death)
			return nil
		}))
	})
	return deaths
}()

func init() {
	http.HandleFunc("/api/deaths/hcf", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostJsonPatch[death](r, func(newDeath *death, r *http.Request) error { return json.NewDecoder(r.Body).Decode(&newDeath) }, &deathsMu, &deaths)
		// TODO -> get and re-render server
	})
	http.HandleFunc("/api/death/mz", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostJsonPatch[death](r, func(newDeath *death, r *http.Request) error { return json.NewDecoder(r.Body).Decode(&newDeath) }, &deathsMu, &deaths)
		// TODO -> get and re-render server
	})
}

type event struct {
	RowNumber int

	StartTimestamp       time.Time `json:"startTimestamp"`
	LootFactor           int       `json:"lootFactor"`
	MaxTimer             int       `json:"maxTimer"`
	IsMovementRestricted bool      `json:"isMovementRestricted"`
	CappingUserUUID      *string   `json:"cappingUserUUID"`
	EndTimestamp         time.Time `json:"endTimestamp"`
	CappingPartyUUID     *string   `json:"cappingPartyUUID"`
	CapMessage           *string   `json:"capMessage"`
	World                string    `json:"world"`
	X                    int       `json:"x"`
	Y                    int       `json:"y"`
	Z                    int       `json:"z"`
	ServerName           string    `json:"serverName"`
	ArenaName            string    `json:"arenaName"`
	Creator              string    `json:"creator"`
}

var events = func() []event {
	var events []event
	getRowsBlocking("SELECT * FROM get_14_newest_network_koths()", func(rows pgx.Rows) {
		var event event
		handleFatalPgx(pgx.ForEachRow(rows, []any{&event.RowNumber, &event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
			events = append(events, event)
			return nil
		}))
	})
	return events
}()

func init() {
	http.HandleFunc("/api/events/hcf", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostJsonPatch[event](r, func(newDeath *event, r *http.Request) error { return json.NewDecoder(r.Body).Decode(&newDeath) }, &eventsMu, &events)
		// TODO ->
	})
	http.HandleFunc("/api/events/mz", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostJsonPatch[event](r, func(newDeath *event, r *http.Request) error { return json.NewDecoder(r.Body).Decode(&newDeath) }, &eventsMu, &events)
		// TODO ->
	})
}

var eventsMu sync.RWMutex

type faction struct {
	name                    string
	partyUuid               string
	frozenUntil             time.Time
	currentMaxDtr           float32
	currentRegenAdjustedDtr float32
}
type bandit struct {
	userUuid            string
	deathId             int
	timestamp           time.Time
	expirationTimestamp time.Time
	deathMessage        string
	deathWorld          string
	deathX              int
	deathY              int
	deathZ              int
}
type serverData struct {
	deathBanMinutes               int
	worldBorderRadius             int
	sharpnessLimit                int
	powerLimit                    int
	protectionLimit               int
	regenLimit                    int
	strengthLimit                 int
	isWeaknessEnabled             bool
	isBardPassiveDebuffingEnabled bool
	dtrFreezeTimer                int
	dtrMax                        float32
	dtrMaxTime                    int
	dtrOffPeakFreezeTime          int
	offPeakLivesNeededAsCents     int
	bardRadius                    int
	rogueRadius                   int
	timestamp                     time.Time
	serverName                    string
	attackSpeedName               string

	currentPlayers []string
	deaths         []death
	events         []event
	donations      []order // TODO impl
	messages       []string
	videos         []string

	factions []faction
	bandits  []bandit
}

var currentHcfServerName string
var serverDatas = func() map[string]*serverData {
	serverDatas := make(map[string]*serverData)
	getRowsBlocking("SELECT * FROM get_server_datas()", func(rows pgx.Rows) {
		var serverDataBuffer serverData
		handleFatalPgx(pgx.ForEachRow(rows, []any{&serverDataBuffer.deathBanMinutes, &serverDataBuffer.worldBorderRadius, &serverDataBuffer.sharpnessLimit, &serverDataBuffer.powerLimit, &serverDataBuffer.protectionLimit, &serverDataBuffer.regenLimit, &serverDataBuffer.strengthLimit, &serverDataBuffer.isWeaknessEnabled, &serverDataBuffer.isBardPassiveDebuffingEnabled, &serverDataBuffer.dtrFreezeTimer, &serverDataBuffer.dtrMax, &serverDataBuffer.dtrMaxTime, &serverDataBuffer.dtrOffPeakFreezeTime, &serverDataBuffer.offPeakLivesNeededAsCents, &serverDataBuffer.bardRadius, &serverDataBuffer.rogueRadius, &serverDataBuffer.timestamp, &serverDataBuffer.serverName, &serverDataBuffer.attackSpeedName}, func() error {
			serverData := serverDataBuffer
			serverDatas[serverDataBuffer.serverName] = &serverData
			return nil
		}))
	})
	var currentPotentialHcfServerTimestamp time.Time
	for _, data := range serverDatas {
		if strings.Contains(data.serverName, "hcf") && data.timestamp.After(currentPotentialHcfServerTimestamp) {
			currentHcfServerName = data.serverName
			currentPotentialHcfServerTimestamp = data.timestamp
		}
	}

	return serverDatas
}()

func init() {
	http.HandleFunc("/api/servers/", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
}

var currentPlayers = func() []string {
	var currentPlayers []string
	getRowsBlocking("SELECT * FROM get_online_players()", func(rows pgx.Rows) {
		var playerName string
		var serverName string
		handleFatalPgx(pgx.ForEachRow(rows, []any{&playerName, &serverName}, func() error {
			currentPlayers = append(currentPlayers, playerName)
			serverDatas[serverName].currentPlayers = append(serverDatas[serverName].currentPlayers, playerName)
			return nil
		}))
		// TODO sort names
	})
	return currentPlayers
}()

func init() {
	http.HandleFunc("/api/online/hcf", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
	http.HandleFunc("/api/online/mz", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	for serverName, serverData := range serverDatas {
		getRowsBlocking("SELECT * FROM get_12_latest_server_deaths($1)", func(rows pgx.Rows) {
			var death death
			handleFatalPgx(pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
				serverData.deaths = append(serverData.deaths, death)
				return nil
			}))
		}, serverName)
		getRowsBlocking("SELECT * FROM get_14_newest_server_koths($1)", func(rows pgx.Rows) {
			var event event
			handleFatalPgx(pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
				serverData.events = append(serverData.events, event)
				return nil
			}))
		}, serverName)
		getRowsBlocking("SELECT * FROM get_7_factions($1)", func(rows pgx.Rows) {
			var faction faction
			handleFatalPgx(pgx.ForEachRow(rows, []any{&faction.name, &faction.partyUuid, &faction.frozenUntil, &faction.currentMaxDtr, &faction.currentRegenAdjustedDtr}, func() error {
				serverData.factions = append(serverData.factions, faction)
				return nil
			}))
		}, serverName)
		getRowsBlocking("SELECT * FROM get_7_newest_bandits($1)", func(rows pgx.Rows) {
			var bandit bandit
			handleFatalPgx(pgx.ForEachRow(rows, []any{&bandit.userUuid, &bandit.deathId, &bandit.timestamp, &bandit.expirationTimestamp, &bandit.deathMessage, &bandit.deathWorld, &bandit.deathX, &bandit.deathY, &bandit.deathZ}, func() error {
				serverData.bandits = append(serverData.bandits, bandit)
				return nil
			}))
		}, serverName)
	}
}

type money struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}
type order struct {
	ID string `json:"id"`
	//		LocationID                string       `json:"location_id"`
	LineItems []struct {
		UID                      string `json:"uid"`
		Quantity                 string `json:"quantity"`
		Name                     string `json:"name"`
		BasePriceMoney           money  `json:"base_price_money"`
		GrossSalesMoney          money  `json:"gross_sales_money"`
		TotalTaxMoney            money  `json:"total_tax_money"`
		TotalDiscountMoney       money  `json:"total_discount_money"`
		TotalMoney               money  `json:"total_money"`
		VariationTotalPriceMoney money  `json:"variation_total_price_money"`
		ItemType                 string `json:"item_type"`
		TotalServiceChargeMoney  money  `json:"total_service_charge_money"`
	} `json:"line_items"`
	//		Fulfillments              []Fulfillment `json:"fulfillments"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	State     string `json:"state"`
	//		Version                   int          `json:"version"`
	TotalTaxMoney money `json:"total_tax_money"`
	//		TotalDiscountMoney        money        `json:"total_discount_money"`
	TotalTipMoney           money `json:"total_tip_money"`
	TotalMoney              money `json:"total_money"`
	TotalServiceChargeMoney money `json:"total_service_charge_money"`
	//		NetAmounts                NetAmounts   `json:"net_amounts"`
	//		Source                    Source       `json:"source"`
	NetAmountDueMoney money `json:"net_amount_due_money"`
}
type payment struct {
	//		ID             string         `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AmountMoney money     `json:"amount_money"`
	TipMoney    money     `json:"tip_money"`
	Status      string    `json:"status"`
	//		DelayDuration  string         `json:"delay_duration"`
	//		SourceType     string         `json:"source_type"`
	//		CardDetails    CardDetails    `json:"card_details"`
	//		LocationID     string         `json:"location_id"`
	OrderID   string   `json:"order_id"`
	RefundIDs []string `json:"refund_ids"`
	//		RiskEvaluation RiskEvaluation `json:"risk_evaluation"`
	//				ProcessingFee  []ProcessingFee `json:"processing_fee"`
	BuyerEmail string `json:"buyer_email_address"`
	//		BillingAddress Address        `json:"billing_address"`
	//		ShippingAddress Address       `json:"shipping_address"`
	//		CustomerID     string         `json:"customer_id"`
	TotalMoney    money `json:"total_money"`
	ApprovedMoney money `json:"approved_money"`
	//		ReceiptNumber  string         `json:"receipt_number"`
	ReceiptURL string `json:"receipt_url"`
	//		DelayAction    string         `json:"delay_action"`
	//		DelayedUntil   time.Time      `json:"delayed_until"`
	//		ApplicationDetails ApplicationDetails `json:"application_details"`
	//		VersionToken   string         `json:"version_token"`
}

var donations = func() []order {
	req := getFatalRequest("GET", "https://connect.squareup.com/v2/payments?location_id="+os.Getenv("SQUARE_LOCATION_ID"), nil)
	addSquareHeaders(req)

	var paymentIds []string
	for _, payment := range handleGetFatalJsonT[struct {
		Payments []payment `json:"payments"`
	}](req).Payments {
		paymentIds = append(paymentIds, payment.OrderID)
	}

	requestJson, err := json.Marshal(struct {
		LocationID string   `json:"location_id"`
		OrderIDs   []string `json:"order_ids"`
	}{
		LocationID: os.Getenv("SQUARE_LOCATION_ID"),
		OrderIDs:   paymentIds,
	})
	handleFatalErr(err)
	req = getFatalRequest("POST", "https://connect.squareup.com/v2/orders/batch-retrieve", bytes.NewBuffer(requestJson))
	addSquareHeaders(req) // shirley this keeps their correct order

	orders := handleGetFatalJsonT[struct {
		Orders []order `json:"orders"`
	}](req).Orders
	var orderIds []string
	for _, orderId := range orders {
		orderIds = append(orderIds, orderId.ID)
	}
	getRowsBlocking("SELECT * FROM get_unsuccessful_transactions($1)", func(rows pgx.Rows) {
		var missedOrderId string
		handleFatalPgx(pgx.ForEachRow(rows, []any{&missedOrderId}, func() error {
			// TODO -> handle missed transactions
			return nil
		}))
	}, orderIds)

	return orders
}()

func init() {
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
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				var potentialFinalRequest donateRequest
				for _, data := range lineItemDatas {
					if data.GamemodeName+"-"+data.ItemName == request.LineItemName {
						potentialFinalRequest.LineItemName = request.LineItemName
						if data.IsPlural {
							potentialFinalRequest.LineItemAmount = int(math.Max(float64(request.LineItemAmount), 1))
						} else {
							potentialFinalRequest.LineItemAmount = 1 // TODO -> allow 0 amount
						}
						potentialFinalRequest.lineItemCostInCents = data.ItemPriceInCents

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
									potentialFinalRequest.LineItemName = request.Username
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
					}
				}
			}()
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
	http.HandleFunc("/api/donations/webhook", func(w http.ResponseWriter, r *http.Request) { // square's payment.create webhook
		if !strings.Contains(r.Header.Get("Authorization"), os.Getenv("SQUARE_ACCESS_TOKEN")) {
			log.Println("err: square webhook auth")
			return
		} else {
			payment := handleGetFatalJsonT[struct {
				//				NotificationURL string  `json:"notification_url"`
				//				StatusCode      int     `json:"status_code"`
				//				PassesFilter    bool    `json:"passes_filter"`
				Payload struct {
					//					MerchantID string `json:"merchant_id"`
					//							Type       string `json:"type"`
					//							EventID    string `json:"event_id"`
					//							CreatedAt  string `json:"created_at"`
					Data struct {
						//		Type   string `json:"type"`
						//		ID     string `json:"id"`
						Object struct {
							Payment payment `json:"payment"`
						} `json:"object"`
					} `json:"data"`
				} `json:"payload"`
			}](r).Payload.Data.Object.Payment

			req := getFatalRequest("GET", "https://connect.squareup.com/v2/orders/"+payment.OrderID, nil)
			defer req.Body.Close()
			order := handleGetFatalJsonT[order](r)

			for _, lineItem := range order.LineItems {
				go func() {
					parts := strings.Split(lineItem.ItemType, ",")
					var bodyJson struct {
						//							Name string `json:"name"`
						ID string `json:"id"`
					}
					for {
						resp, err := getMojangApiUuidRequest(parts[1])
						if err != nil {
							log.Println(err)
						} else {
							err = json.NewDecoder(resp.Body).Decode(&bodyJson)
							resp.Body.Close()
							if err != nil {
								log.Fatal(err)
							} else {
								break
							}
						}
					}
					_, err := postgresPool.Exec(context.Background(), "CALL insert_successful_transaction($1, $2, $3, $4, $5, $6)", payment.OrderID, bodyJson.ID, parts[0], parts[1], lineItem.Quantity, payment.AmountMoney.Amount-payment.TipMoney.Amount)
					if err != nil {
						log.Fatal(err)
					}
				}()
			}
		}
	})
}

type discordMessage struct {
	Type         int           `json:"type"`
	Content      string        `json:"content"`
	Mentions     []interface{} `json:"mentions"`
	MentionRoles []interface{} `json:"mention_roles"`
	Attachments  []struct {
		ID          string `json:"id"`
		Filename    string `json:"filename"`
		Size        int    `json:"size"`
		URL         string `json:"url"`
		ProxyURL    string `json:"proxy_url"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		ContentType string `json:"content_type"`
	} `json:"attachments"`
	Embeds          []interface{} `json:"embeds"`
	Timestamp       string        `json:"timestamp"`
	EditedTimestamp interface{}   `json:"edited_timestamp"`
	Flags           int           `json:"flags"`
	Components      []interface{} `json:"components"`
	ID              string        `json:"id"`
	ChannelID       string        `json:"channel_id"`
	Author          struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Avatar        string `json:"avatar"`
		Discriminator string `json:"discriminator"`
		GlobalName    string `json:"global_name"`
	} `json:"author"`
	Pinned          bool `json:"pinned"`
	MentionEveryone bool `json:"mention_everyone"`
	Reactions       []struct {
		Emoji struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"emoji"`
		Count int `json:"count"`
	} `json:"reactions"`
}

func getDiscordMessages(channelId string) []discordMessage {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/channels/"+channelId+"/messages?limit=50", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bot "+os.Getenv("DISCORD_BOT_TOKEN"))
	return handleGetFatalJsonT[[]discordMessage](req)
}

var discordMessages = getDiscordMessages("1245300045188956255")
var changelog = getDiscordMessages("1346008874830008375")
var announcements = func() []discordMessage {
	allAnnouncementsMessages := getDiscordMessages("1265836245678948464")
	var importantAnnouncementsMessages []discordMessage
	for _, discordMessage := range allAnnouncementsMessages {
		if discordMessage.MentionEveryone {
			importantAnnouncementsMessages = append(importantAnnouncementsMessages, discordMessage)
		}
	}
	return importantAnnouncementsMessages
}()                   // TODO -> store last checked time and then check for every join or something + refresh button + reddit too
var messages []string // TODO -> make player name clickable, maybe velocity/paper will let me query message history
func init() {
	http.HandleFunc("/api/chat/hcf", func(w http.ResponseWriter, r *http.Request) {
		// TODO messages + ServerData.Messages
	})
	http.HandleFunc("/api/chat/mz", func(w http.ResponseWriter, r *http.Request) {
		// TODO messages + ServerData.Messages
	})
}

type lineItemData struct {
	GamemodeName     string
	ItemName         string
	ItemPriceInCents int
	ItemDescription  string
	IsPlural         bool

	ItemPriceInDollars int
}

var lineItemDatas = func() []lineItemData {
	var slice []lineItemData
	getRowsBlocking("SELECT * FROM get_line_items()", func(rows pgx.Rows) {
		var death lineItemData
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.GamemodeName, &death.ItemName, &death.ItemPriceInCents, &death.ItemDescription, &death.IsPlural}, func() error {
			death.ItemPriceInDollars = death.ItemPriceInCents / 100.0
			slice = append(slice, death)
			return nil
		}))
	})
	return slice
}()

func getMainTemplate(fileName string) *template.Template {
	mainTemplate, err := template.ParseFiles("main.html", fileName)
	handleFatalErr(err)
	return mainTemplate
}

var homeTemplate = getMainTemplate("main-home.html")
var mzTemplate = getMainTemplate("main-mz.html")
var hcfTemplate = getMainTemplate("main-hcf.html")

type mainTemplateData struct {
	NetworkPlayers     []string
	ServerPlayers      []string
	NewPlayers         []newPlayer
	PotpissersTips     []string
	Deaths             []death
	Messages           []string
	Events             []event
	Announcements      []discordMessage
	Changelog          []discordMessage
	DiscordMessages    []discordMessage
	Donations          []order
	OffPeakLivesNeeded float32
	PeakLivesNeeded    float32
	LineItemData       []lineItemData
}

func getHome() []byte {
	var buffer bytes.Buffer
	offPeakLivesNeeded := float32(serverDatas[currentHcfServerName].offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(homeTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData
	}{
		MainTemplateData: mainTemplateData{
			NetworkPlayers:     currentPlayers,
			ServerPlayers:      serverDatas["hub"].currentPlayers,
			NewPlayers:         newPlayers,
			PotpissersTips:     potpissersTips,
			Deaths:             deaths,
			Messages:           messages,
			Events:             events,
			Announcements:      announcements,
			Changelog:          changelog,
			DiscordMessages:    discordMessages,
			Donations:          donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded:    offPeakLivesNeeded / 2,
			LineItemData:       lineItemDatas,
		},
	}))
	return buffer.Bytes()
}
func getMz() []byte {
	var buffer bytes.Buffer
	mzData := serverDatas["mz"]
	offPeakLivesNeeded := float32(serverDatas[currentHcfServerName].offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(mzTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData

		AttackSpeed string

		MzTips  []string
		Bandits []bandit
	}{
		MainTemplateData: mainTemplateData{
			NetworkPlayers:     currentPlayers,
			ServerPlayers:      mzData.currentPlayers,
			NewPlayers:         newPlayers,
			PotpissersTips:     potpissersTips,
			Deaths:             mzData.deaths,
			Messages:           mzData.messages,
			Events:             mzData.events,
			Announcements:      announcements,
			Changelog:          changelog,
			DiscordMessages:    discordMessages,
			Donations:          donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded:    offPeakLivesNeeded / 2,
			LineItemData:       lineItemDatas,
		},

		AttackSpeed: mzData.attackSpeedName,

		MzTips:  mzTips,
		Bandits: mzData.bandits,
	}))
	return buffer.Bytes()
}
func getHcf() []byte {
	var buffer bytes.Buffer
	serverData := serverDatas[currentHcfServerName]
	offPeakLivesNeeded := float32(serverData.offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(hcfTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData

		AttackSpeed string

		DeathBanMinutes int
		LootFactor      int
		BorderSize      int

		SharpnessLimit                int
		ProtectionLimit               int
		PowerLimit                    int
		RegenLimit                    int
		StrengthLimit                 int
		IsWeaknessEnabled             bool
		IsBardPassiveDebuffingEnabled bool
		DtrMax                        float32

		CubecoreTips []string
		ClassTips    []string
		Factions     []faction
	}{
		MainTemplateData: mainTemplateData{
			NetworkPlayers:     currentPlayers,
			ServerPlayers:      serverData.currentPlayers,
			NewPlayers:         newPlayers,
			PotpissersTips:     potpissersTips,
			Deaths:             deaths,
			Messages:           messages,
			Events:             serverData.events,
			Announcements:      announcements,
			Changelog:          changelog,
			DiscordMessages:    discordMessages,
			Donations:          donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded:    offPeakLivesNeeded / 2,
			LineItemData:       lineItemDatas,
		},

		AttackSpeed: serverData.attackSpeedName,

		DeathBanMinutes: serverData.deathBanMinutes,
		//			LootFactor: serverDatas["hcf"]., // TODO -> defaultLootFactor
		BorderSize: serverData.worldBorderRadius,

		SharpnessLimit:                serverData.sharpnessLimit,
		ProtectionLimit:               serverData.protectionLimit,
		PowerLimit:                    serverData.powerLimit,
		RegenLimit:                    serverData.regenLimit,
		StrengthLimit:                 serverData.strengthLimit,
		IsWeaknessEnabled:             serverData.isWeaknessEnabled,
		IsBardPassiveDebuffingEnabled: serverData.isBardPassiveDebuffingEnabled,
		DtrMax:                        serverData.dtrMax,

		CubecoreTips: cubecoreTips,
		ClassTips:    cubecoreClassTips,
		Factions:     serverData.factions,
	}))
	return buffer.Bytes()
}

var home = getHome()
var mz = getMz()
var hcf = getHcf()

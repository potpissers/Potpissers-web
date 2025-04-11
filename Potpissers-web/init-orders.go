package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

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

var donationsMu sync.RWMutex
var donations = func() []order {
	addSquareHeaders := func(request *http.Request) {
		request.Header.Add("Authorization", "Bearer "+os.Getenv("SQUARE_ACCESS_TOKEN"))
		request.Header.Add("Content-Type", "application/json")
	}
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
	addSquareHeaders(req) // TODO shirley this keeps their correct order

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

	println("orders done")

	return orders
}()

func init() {
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
			newOrder := handleGetFatalJsonT[order](r)

			for _, lineItem := range newOrder.LineItems {
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
			donationsMu.Lock()
			donations = append([]order{newOrder}, donations...)
			donationsMu.Unlock()
			jsonBytes, err := json.Marshal(sseMessage{"donations", newOrder})
			if err != nil {
				log.Fatal(err)
			}
			home = getMainTemplateBytes("hub")
			mz = getMainTemplateBytes("mz")
			hcf = getMainTemplateBytes("hcf" + currentHcfServerName)
			handleSseData(jsonBytes, mainConnections)
		}
	})
}

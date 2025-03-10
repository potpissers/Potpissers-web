package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var potpissersTips = getTipsBlocking("null")
var cubecoreTips = getTipsBlocking("cubecore")
var cubecoreClassTips = getTipsBlocking("cubecore_classes")
var mzTips = getTipsBlocking("minez")
type newPlayer struct {
	PlayerUuid string `json:"playerUuid"`
	Referrer  string `json:"referrer"`
	Timestamp time.Time `json:"timestamp"`
	RowNumber int `json:"rowNumber"`
}
var newPlayers = func() []newPlayer {
	var newPlayers []newPlayer
	getRowsBlocking(Return12NewPlayers, func(rows pgx.Rows) {
		var death newPlayer
		_, err := pgx.ForEachRow(rows, []any{&death.PlayerUuid, &death.Referrer, &death.Timestamp, &death.RowNumber}, func() error {
			newPlayers = append(newPlayers, death)
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	return newPlayers
}()
func init() {
	http.HandleFunc("/api/new-players", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[newPlayer](r, func(newT *newPlayer, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newT)}, &newPlayersMu, &newPlayers)
	})
}
var newPlayersMu sync.RWMutex
type death struct {
	ServerName string `json:"serverName"`
	VictimUserFightId *int `json:"victimUserFightId"`
	Timestamp time.Time `json:"timestamp"`
	VictimUuid string `json:"victimUuid"`
	// TODO victim inventory
	DeathWorldName string `json:"deathWorldName"`
	DeathX int `json:"deathX"`
	DeathY int `json:"deathY"`
	DeathZ int `json:"deathZ"`
	DeathMessage string `json:"deathMessage"`
	KillerUuid *string `json:"killerUuid"`
	// TODO killer weapon
	// TODO killer inventory
}
var deaths = func() []death {
	var deaths []death
	getRowsBlocking(Return12Deaths, func(rows pgx.Rows) {
		var death death
		_, err := pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
			deaths = append(deaths, death)
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	return deaths
}()
func init() {
	http.HandleFunc("/api/deaths", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[death](r, func(newDeath *death, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &deathsMu, &deaths)
	})
}
var deathsMu sync.RWMutex
type event struct {
	StartTimestamp time.Time `json:"startTimestamp"`
	LootFactor    int `json:"lootFactor"`
	MaxTimer int `json:"maxTimer"`
	IsMovementRestricted bool `json:"isMovementRestricted"`
	CappingUserUUID *string `json:"cappingUserUUID"`
	EndTimestamp time.Time `json:"endTimestamp"`
	CappingPartyUUID *string `json:"cappingPartyUUID"`
	CapMessage *string `json:"capMessage"`
	World string `json:"world"`
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
	ServerName string `json:"serverName"`
	ArenaName string `json:"arenaName"`
	Creator string `json:"creator"`
}
var events = func() []event {
	var events []event
	getRowsBlocking(Return14Events, func(rows pgx.Rows) {
		var event event
		_, err := pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
			events = append(events, event)
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	return events
}()
func init() {
	http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[event](r, func(newDeath *event, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &eventsMu, &events)
	})
}
var eventsMu sync.RWMutex
type faction struct {
	Name string
	PartyUuid string
}
type bandit struct {
	UserUuid string
	DeathId int
	Timestamp time.Time
	ExpirationTimestamp time.Time
	DeathMessage string
	DeathWorld string
	DeathX int
	DeathY int
	DeathZ int
}
type serverData struct {
	DeathBanMinutes int
	WorldBorderRadius int
	SharpnessLimit int
	PowerLimit int
	ProtectionLimit int
	RegenLimit int
	StrengthLimit int
	IsWeaknessEnabled bool
	IsBardPassiveDebuffingEnabled bool
	DtrFreezeTimer int
	DtrMax float32
	DtrMaxTime int
	DtrOffPeakFreezeTime int
	OffPeakLivesNeededAsCents int
	BardRadius int
	RogueRadius int
	ServerName string
	AttackSpeedName string

	CurrentPlayers []string
	Deaths []death
	Events []event
	//		Transaction []Transaction TODO
	Messages []string
	Videos []string

	Factions []faction
	Bandits []bandit
}
var serverDatas = func() map[string]*serverData {
	serverDatas := make(map[string]*serverData)
	getRowsBlocking(ReturnAllServerData, func(rows pgx.Rows) {
		var serverData serverData
		_, err := pgx.ForEachRow(rows, []any{&serverData.DeathBanMinutes, &serverData.WorldBorderRadius, &serverData.SharpnessLimit, &serverData.PowerLimit, &serverData.ProtectionLimit, &serverData.RegenLimit, &serverData.StrengthLimit, &serverData.IsWeaknessEnabled, &serverData.IsBardPassiveDebuffingEnabled, &serverData.DtrFreezeTimer, &serverData.DtrMax, &serverData.DtrMaxTime, &serverData.DtrOffPeakFreezeTime, &serverData.OffPeakLivesNeededAsCents, &serverData.BardRadius, &serverData.RogueRadius, &serverData.ServerName, &serverData.AttackSpeedName}, func() error {
			serverDatas[serverData.ServerName] = &serverData
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	return serverDatas
}()
func init() {
	http.HandleFunc("/api/servers/", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
}
var currentPlayers = func() []string  {
	var currentPlayers []string
	getRowsBlocking(ReturnAllOnlinePlayers, func(rows pgx.Rows) {
		var playerName string
		var serverName string
		_, err := pgx.ForEachRow(rows, []any{&playerName, &serverName}, func() error {
			currentPlayers = append(currentPlayers, playerName)
			serverDatas[serverName].CurrentPlayers = append(serverDatas[serverName].CurrentPlayers, playerName)
			return nil
		})
		// TODO sort names
		if err != nil {
			log.Fatal(err)
		}
	})
	return currentPlayers
}()
func init() {
	for serverName, serverData := range serverDatas {
		getRowsBlocking(Return12ServerDeaths, func(rows pgx.Rows) {
			var death death
			_, err := pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
				serverData.Deaths = append(serverData.Deaths, death)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
		getRowsBlocking(Return14ServerEvents, func(rows pgx.Rows) {
			var event event
			_, err := pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
				serverData.Events = append(serverData.Events, event)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
		getRowsBlocking(Return7ServerFactions, func(rows pgx.Rows) {
			var faction faction
			_, err := pgx.ForEachRow(rows, []any{&faction.Name, &faction.PartyUuid}, func() error {
				serverData.Factions = append(serverData.Factions, faction)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
		getRowsBlocking(Return7ServerBandits, func(rows pgx.Rows) {
			var bandit bandit
			_, err := pgx.ForEachRow(rows, []any{&bandit.UserUuid, &bandit.DeathId, &bandit.Timestamp, &bandit.ExpirationTimestamp, &bandit.DeathMessage, &bandit.DeathWorld, &bandit.DeathX, &bandit.DeathY, &bandit.DeathZ}, func() error {
				serverData.Bandits = append(serverData.Bandits, bandit)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
	}
}
type Money struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}
type LineItem struct {
	UID                        string   `json:"uid"`
	Quantity                   string   `json:"quantity"`
	Name                       string   `json:"name"`
	BasePriceMoney             Money   `json:"base_price_money"`
	GrossSalesMoney            Money   `json:"gross_sales_money"`
	TotalTaxMoney              Money   `json:"total_tax_money"`
	TotalDiscountMoney         Money   `json:"total_discount_money"`
	TotalMoney                 Money   `json:"total_money"`
	VariationTotalPriceMoney   Money   `json:"variation_total_price_money"`
	ItemType                   string  `json:"item_type"`
	TotalServiceChargeMoney    Money   `json:"total_service_charge_money"`
}
type Order struct {
	ID                        string       `json:"id"`
	//		LocationID                string       `json:"location_id"`
	LineItems                 []LineItem   `json:"line_items"`
	//		Fulfillments              []Fulfillment `json:"fulfillments"`
	CreatedAt                 string       `json:"created_at"`
	UpdatedAt                 string       `json:"updated_at"`
	State                     string       `json:"state"`
	//		Version                   int          `json:"version"`
	TotalTaxMoney             Money        `json:"total_tax_money"`
	//		TotalDiscountMoney        Money        `json:"total_discount_money"`
	TotalTipMoney             Money        `json:"total_tip_money"`
	TotalMoney                Money        `json:"total_money"`
	TotalServiceChargeMoney   Money        `json:"total_service_charge_money"`
	//		NetAmounts                NetAmounts   `json:"net_amounts"`
	//		Source                    Source       `json:"source"`
	NetAmountDueMoney         Money        `json:"net_amount_due_money"`
}
type Payment struct {
	//		ID             string         `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	AmountMoney    Money          `json:"amount_money"`
	TipMoney       Money          `json:"tip_money"`
	Status         string         `json:"status"`
	//		DelayDuration  string         `json:"delay_duration"`
	//		SourceType     string         `json:"source_type"`
	//		CardDetails    CardDetails    `json:"card_details"`
	//		LocationID     string         `json:"location_id"`
	OrderID        string         `json:"order_id"`
	RefundIDs      []string       `json:"refund_ids"`
	//		RiskEvaluation RiskEvaluation `json:"risk_evaluation"`
	//				ProcessingFee  []ProcessingFee `json:"processing_fee"`
	BuyerEmail     string         `json:"buyer_email_address"`
	//		BillingAddress Address        `json:"billing_address"`
	//		ShippingAddress Address       `json:"shipping_address"`
	//		CustomerID     string         `json:"customer_id"`
	TotalMoney     Money          `json:"total_money"`
	ApprovedMoney  Money          `json:"approved_money"`
	//		ReceiptNumber  string         `json:"receipt_number"`
	ReceiptURL     string         `json:"receipt_url"`
	//		DelayAction    string         `json:"delay_action"`
	//		DelayedUntil   time.Time      `json:"delayed_until"`
	//		ApplicationDetails ApplicationDetails `json:"application_details"`
	//		VersionToken   string         `json:"version_token"`
}
var donations = func() []Order {
	req := getFatalRequest("https://connect.squareup.com/v2/payments?location_id=" + os.Getenv("SQUARE_LOCATION_ID"), nil)
	addSquareHeaders := func(request *http.Request) {
		request.Header.Add("Authorization", "Bearer " + os.Getenv("SQUARE_ACCESS_TOKEN"))
		request.Header.Add("Content-Type", "application/json")
	}
	addSquareHeaders(req)

	type PaymentResponse struct {
		Payments []Payment `json:"payments"`
	}
	var paymentIds []string
	for _, payment := range getJsonT[PaymentResponse](req).Payments {
		paymentIds = append(paymentIds, payment.OrderID)
	}

	requestJson, err := json.Marshal(struct {
		LocationID string   `json:"location_id"`
		OrderIDs   []string `json:"order_ids"`
	} {
		LocationID: os.Getenv("SQUARE_LOCATION_ID"),
		OrderIDs: paymentIds,
		})
	if err != nil {
		log.Fatal(err)
	}
	req = getFatalRequest("https://connect.squareup.com/v2/orders/batch-retrieve", bytes.NewBuffer(requestJson))
	addSquareHeaders(req) // shirley this keeps their correct order
	type OrderResponse struct {
		Orders []Order `json:"orders"`
	}

	orders := getJsonT[OrderResponse](req).Orders
	var orderIds []string
	for _, orderId := range orders {
		orderIds = append(orderIds, orderId.ID)
	}
	getRowsBlocking(ReturnUnsuccessfulTransactions, func(rows pgx.Rows) {
		var missedOrderId string
		_, err := pgx.ForEachRow(rows, []any{&missedOrderId}, func() error {
			// TODO -> handle missed transactions
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}, orderIds)

	return orders
}()
func init() {
	http.HandleFunc("/api/donations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET": { // square's payment.create
			type Object struct {
				Payment Payment `json:"payment"`
			}
			type Data struct {
				//		Type   string `json:"type"`
				//		ID     string `json:"id"`
				Object Object `json:"object"`
			}
			type Payload struct {
				//		MerchantID string `json:"merchant_id"`
				//		Type       string `json:"type"`
				//		EventID    string `json:"event_id"`
				//		CreatedAt  string `json:"created_at"`
				Data       Data   `json:"data"`
			}
			type Notification struct {
				//		NotificationURL string  `json:"notification_url"`
				//		StatusCode      int     `json:"status_code"`
				//		PassesFilter    bool    `json:"passes_filter"`
				Payload         Payload `json:"payload"`
			}
			payment := getJsonT[Notification](r).Payload.Data.Object.Payment
		}
		case "POST": { // TODO GET ?
			{ // TODO GET ?
				var donationRequest []struct {
					Username string `json:"username"`
					LineItemName string `json:"line_item_name"`
					LineItemAmount int `json:"line_item_amount"`

					LineItemCostInCents int
				}
				err := json.NewDecoder(r.Body).Decode(&donationRequest)
				if err != nil {
					log.Println(err)
					return
				}
				outer:
					for i := range donationRequest {
						for _, data := range lineItemData {
							if data.ServerName + "-" + data.ItemName == donationRequest[i].LineItemName {
								if data.IsPlural {
									donationRequest[i].LineItemAmount = int(math.Max(float64(donationRequest[i].LineItemAmount), 1))
								} else {
									donationRequest[i].LineItemAmount = 1 // TODO -> maybe allow 0 amount
								}
								donationRequest[i].LineItemCostInCents = data.ItemPriceInCents
								continue outer
							}
						}
						// else
						log.Println("error: invalid donationRequest line item")
						return // door nigga from game of thrones
					}

					type LineItem struct {
						Quantity       string `json:"quantity"`
						ItemType       string `json:"item_type"`
						Name           string `json:"name"`
						BasePriceMoney Money `json:"base_price_money"`
					}
					var lineItems []LineItem
					for _, lineItem := range donationRequest {
						lineItems = append(lineItems, LineItem {
							Quantity: strconv.Itoa(lineItem.LineItemAmount),
							ItemType: "ITEM",
							Name: lineItem.LineItemName + "," + lineItem.Username,
							BasePriceMoney: Money {
								Amount: lineItem.LineItemCostInCents,
								Currency: "USD",
								},
								})
					}

					type Source struct {
						Name string `json:"name"`
					}
					type Fulfillment struct {
						UID  string `json:"uid"`
						Type string `json:"type"`
						State string `json:"state"`
					}
					type NetAmounts struct {
						TotalMoney           Money `json:"total_money"`
						TaxMoney             Money `json:"tax_money"`
						DiscountMoney        Money `json:"discount_money"`
						TipMoney             Money `json:"tip_money"`
						ServiceChargeMoney   Money `json:"service_charge_money"`
					}
					type Order struct {
						LocationID string `json:"location_id"`
						LineItems  []LineItem `json:"line_items"`
					}
					type AcceptedPaymentMethods struct {
						AfterpayClearpay bool `json:"afterpay_clearpay"`
						ApplePay         bool `json:"apple_pay"`
						CashAppPay       bool `json:"cash_app_pay"`
						GooglePay        bool `json:"google_pay"`
					}
					type CheckoutOptions struct {
						AllowTipping          bool `json:"allow_tipping"`
						AcceptedPaymentMethods AcceptedPaymentMethods `json:"accepted_payment_methods"`
						AskForShippingAddress bool   `json:"ask_for_shipping_address"`
						EnableCoupon          bool   `json:"enable_coupon"`
						EnableLoyalty         bool   `json:"enable_loyalty"`
						MerchantSupportEmail  string `json:"merchant_support_email"`
						RedirectURL           string `json:"redirect_url"`
					}
					reqData := struct {
						CheckoutOptions CheckoutOptions `json:"checkout_options"`
						Description string `json:"description"`
						Order     Order   `json:"order"`
					} {
						CheckoutOptions: CheckoutOptions {
							AllowTipping: true,
							AcceptedPaymentMethods: AcceptedPaymentMethods{
								AfterpayClearpay: false,
								ApplePay: true,
								CashAppPay: true,
								GooglePay: true,
								},
								AskForShippingAddress: false,
								EnableCoupon: false,
								EnableLoyalty: false,
								MerchantSupportEmail: "potpissers@gmail.com",
								RedirectURL: "potpissers.com/donations",
								},
								Description: "hey",
								Order: Order {
							LocationID: os.Getenv("SQUARE_LOCATION_ID"),
							LineItems: lineItems,
							},
							}
							reqBody, err := json.Marshal(reqData)
							if err != nil {
								log.Fatal(err)
							}
							req, err := http.NewRequest("POST", "https://connect.squareup.com/v2/online-checkout/payment-links", bytes.NewBuffer(reqBody))
							if err != nil {
								log.Fatal(err)
							}
							req.Header.Set("Square-Version", "2025-02-20")
							req.Header.Set("Authorization", "Bearer " + os.Getenv("SQUARE_ACCESS_TOKEN"))
							req.Header.Set("Content-Type", "application/json")

							resp, err := (&http.Client{}).Do(req)
							if err != nil {
								log.Fatal(err)
							}
							defer func(Body io.ReadCloser) {
								err := Body.Close()
								if err != nil {
									log.Fatal(err)
								}
							}(resp.Body)

							type LineItemResponse struct {
								Quantity       string `json:"quantity"`
								ItemType       string `json:"item_type"`
								Name           string `json:"name"`
								BasePriceMoney Money `json:"base_price_money"`

								UID                      string `json:"uid"`
								VariationTotalPriceMoney Money  `json:"variation_total_price_money"`
								GrossSalesMoney          Money  `json:"gross_sales_money"`
								TotalTaxMoney            Money  `json:"total_tax_money"`
								TotalDiscountMoney       Money  `json:"total_discount_money"`
								TotalMoney               Money  `json:"total_money"`
								TotalServiceChargeMoney  Money  `json:"total_service_charge_money"`
							}
							type OrderResponse struct {
								LocationID string `json:"location_id"`
								LineItems  []LineItemResponse `json:"line_items"`

								ID                    string        `json:"id"`
								Source                Source        `json:"source"`
								Fulfillments          []Fulfillment `json:"fulfillments"`
								NetAmounts            NetAmounts    `json:"net_amounts"`
								CreatedAt             time.Time     `json:"created_at"`
								UpdatedAt             time.Time     `json:"updated_at"`
								State                 string        `json:"state"`
								Version               int           `json:"version"`
								TotalMoney            Money         `json:"total_money"`
								TotalTaxMoney         Money         `json:"total_tax_money"`
								TotalDiscountMoney    Money         `json:"total_discount_money"`
								TotalTipMoney         Money         `json:"total_tip_money"`
								TotalServiceChargeMoney Money       `json:"total_service_charge_money"`
								NetAmountDueMoney     Money         `json:"net_amount_due_money"`
							}
							type RelatedResources struct {
								Orders []OrderResponse `json:"orders"`
							}
							type PaymentLink struct {
								ID                 string         `json:"id"`
								Version           int            `json:"version"`
								Description       string         `json:"description"`
								OrderID           string         `json:"order_id"`
								CheckoutOptions   CheckoutOptions `json:"checkout_options"`
								URL               string         `json:"url"`
								LongURL           string         `json:"long_url"`
								CreatedAt         time.Time      `json:"created_at"`
							}
							var paymentLinkResp struct {
								PaymentLink      PaymentLink      `json:"payment_link"`
								RelatedResources RelatedResources `json:"related_resources"`
							}
							if err := json.NewDecoder(resp.Body).Decode(&paymentLinkResp); err != nil {
								log.Fatal(err)
							}
							_, err = w.Write([]byte(paymentLinkResp.PaymentLink.URL))
							if err != nil {
								log.Fatal(err)
							}
			}
		}
		}
	})
}
type Author struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name"`
}
type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	ProxyURL    string `json:"proxy_url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ContentType string `json:"content_type"`
}
type Reaction struct {
	Emoji struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"emoji"`
	Count int `json:"count"`
}
type DiscordMessage struct {
	Type           int           `json:"type"`
	Content        string        `json:"content"`
	Mentions       []interface{} `json:"mentions"`
	MentionRoles   []interface{} `json:"mention_roles"`
	Attachments    []Attachment  `json:"attachments"`
	Embeds         []interface{} `json:"embeds"`
	Timestamp      string        `json:"timestamp"`
	EditedTimestamp interface{}   `json:"edited_timestamp"`
	Flags          int           `json:"flags"`
	Components     []interface{} `json:"components"`
	ID             string        `json:"id"`
	ChannelID      string        `json:"channel_id"`
	Author         Author        `json:"author"`
	Pinned         bool          `json:"pinned"`
	MentionEveryone bool         `json:"mention_everyone"`
	Reactions      []Reaction    `json:"reactions"`
}
var discordMessages = getDiscordMessages("1245300045188956255")
var changelog = getDiscordMessages("1346008874830008375")
var announcements = func() []DiscordMessage {
	allAnnouncementsMessages := getDiscordMessages("1265836245678948464")
	var importantAnnouncementsMessages []DiscordMessage
	for _, discordMessage := range allAnnouncementsMessages {
		if discordMessage.MentionEveryone {
			importantAnnouncementsMessages = append(importantAnnouncementsMessages, discordMessage)
		}
	}
	return importantAnnouncementsMessages
}() // TODO -> store last checked time and then check for every join or something + refresh button + reddit too
var messages []string // TODO -> make player name clickable, maybe velocity/paper will let me query message history
func init() {
	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		// TODO messages + ServerData.Messages
	})
}
type LineItemData struct {
	ServerName string
	ItemName string
	ItemPriceInDollars int
	ItemPriceInCents int
	ItemDescription string
	IsPlural bool
}
var lineItemData = func() []LineItemData {
	offPeakLivesNeeded := float32(serverDatas["hcf"].OffPeakLivesNeededAsCents) / 100.0
	lineItemData := []LineItemData {
		{
			ServerName: "hcf",
			ItemName: "life",
			ItemPriceInDollars: 4,
			ItemPriceInCents: 400, // TODO -> db + ingame
			ItemDescription: fmt.Sprintf("/revive (username). removes deathban (alts aren't affected). current revive life cost: %g & %g during events", offPeakLivesNeeded, offPeakLivesNeeded / 2),
			IsPlural: true,
			},
			{
			ServerName: "hcf",
			ItemName: "basic",
			ItemPriceInDollars: 8,
			ItemPriceInCents: 800,
			ItemDescription: "green name, basic server slot, and revive cost + deathban reduced to 80%",
			IsPlural: false,
			},
			{
			ServerName: "hcf",
			ItemName: "gold",
			ItemPriceInDollars: 16,
			ItemPriceInCents: 1600,
			ItemDescription: "yellow name, gold server slot, and revive cost + deathban reduced to 60%",
			IsPlural: false,
			},
			{
			ServerName: "hcf",
			ItemName: "diamond",
			ItemPriceInDollars: 24,
			ItemPriceInCents: 2400,
			ItemDescription: "aqua name, diamond server slot, and revive cost + deathban reduced to 40%",
			IsPlural: false,
			},
			{
			ServerName: "hcf",
			ItemName: "ruby",
			ItemPriceInDollars: 32,
			ItemPriceInCents: 3200,
			ItemDescription: "red name, ruby server slot, and revive cost + deathban reduced to 20%",
			IsPlural: false,
			},

			{
			ServerName: "mz",
			ItemName: "life",
			ItemPriceInDollars: 4,
			ItemPriceInCents: 400,
			ItemDescription: "/revive (username). removes alt deathban",
			IsPlural: true,
			},
			{
			ServerName: "mz",
			ItemName: "basic",
			ItemPriceInDollars: 6,
			ItemPriceInCents: 600,
			ItemDescription: "green name, basic server slot",
			IsPlural: false,
			},
			{
			ServerName: "mz",
			ItemName: "gold",
			ItemPriceInDollars: 12,
			ItemPriceInCents: 1200,
			ItemDescription: "yellow name, gold server slot",
			IsPlural: false,
			},
			{
			ServerName: "mz",
			ItemName: "diamond",
			ItemPriceInDollars: 18,
			ItemPriceInCents: 1800,
			ItemDescription: "aqua name, diamond server slot",
			IsPlural: false,
			},
			{
			ServerName: "mz",
			ItemName: "ruby",
			ItemPriceInDollars: 24,
			ItemPriceInCents: 2400,
			ItemDescription: "red name, ruby server slot",
			IsPlural: false,
			},
			}
			for _, lineItem := range lineItemData {
				_, err := postgresPool.Exec(context.Background(), , lineItem.ItemName, value)
				if err != nil {
					log.Fatal(err)
				}
			} // TODO -> handle this
}()
var homeTemplate = getMainTemplate("main-home.html")
var mzTemplate = getMainTemplate("main-mz.html")
var hcfTemplate = getMainTemplate("main-hcf.html")
type MainTemplateData struct {
	NetworkPlayers []string
	ServerPlayers []string
	NewPlayers []newPlayer
	PotpissersTips []string
	Deaths []death
	Messages []string
	Events []event
	Announcements []DiscordMessage
	Changelog []DiscordMessage
	DiscordMessages []DiscordMessage
	Donations []Order
	OffPeakLivesNeeded float32
	PeakLivesNeeded float32
	LineItemData []LineItemData
}
var home = getHome()
var mz = getMz()
var hcf = getHcf()

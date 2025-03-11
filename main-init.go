package main

import (
	"bytes"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)
func getTipsBlocking(tipsName string) []string {
	var tips []string
	getRowsBlocking(ReturnServerTips, func(rows pgx.Rows) {
		var tipMessage string
		handleFatalPgx(pgx.ForEachRow(rows, []any{&tipMessage}, func() error {
			tips = append(tips, tipMessage)
			return nil
		}))
	}, tipsName)
	return tips
}
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
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.PlayerUuid, &death.Referrer, &death.Timestamp, &death.RowNumber}, func() error {
			newPlayers = append(newPlayers, death)
			return nil
		}))
	})
	return newPlayers
}()
func init() {
	http.HandleFunc("/api/new-players", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostPutJson[newPlayer](r, func(newT *newPlayer, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newT)}, &newPlayersMu, &newPlayers)
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
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
			deaths = append(deaths, death)
			return nil
		}))
	})
	return deaths
}()
func init() {
	http.HandleFunc("/api/deaths", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostPutJson[death](r, func(newDeath *death, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &deathsMu, &deaths)
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
		handleFatalPgx(pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
			events = append(events, event)
			return nil
		}))
	})
	return events
}()
func init() {
	http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		handleLocalhostPutJson[event](r, func(newDeath *event, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &eventsMu, &events)
	})
}
var eventsMu sync.RWMutex
type faction struct {
	name string
	partyUuid string
}
type bandit struct {
	userUuid string
	deathId int
	timestamp time.Time
	expirationTimestamp time.Time
	deathMessage string
	deathWorld string
	deathX int
	deathY int
	deathZ int
}
type serverData struct {
	deathBanMinutes int
	worldBorderRadius int
	sharpnessLimit int
	powerLimit int
	protectionLimit int
	regenLimit int
	strengthLimit int
	isWeaknessEnabled bool
	isBardPassiveDebuffingEnabled bool
	dtrFreezeTimer int
	dtrMax float32
	dtrMaxTime int
	dtrOffPeakFreezeTime int
	offPeakLivesNeededAsCents int
	bardRadius int
	rogueRadius int
	serverName string
	attackSpeedName string

	currentPlayers []string
	deaths []death
	events []event
	//		Transaction []Transaction TODO
	messages []string
	videos []string

	factions []faction
	bandits []bandit
}
var serverDatas = func() map[string]*serverData {
	serverDatas := make(map[string]*serverData)
	getRowsBlocking(ReturnAllServerData, func(rows pgx.Rows) {
		var serverData serverData
		handleFatalPgx(pgx.ForEachRow(rows, []any{&serverData.deathBanMinutes, &serverData.worldBorderRadius, &serverData.sharpnessLimit, &serverData.powerLimit, &serverData.protectionLimit, &serverData.regenLimit, &serverData.strengthLimit, &serverData.isWeaknessEnabled, &serverData.isBardPassiveDebuffingEnabled, &serverData.dtrFreezeTimer, &serverData.dtrMax, &serverData.dtrMaxTime, &serverData.dtrOffPeakFreezeTime, &serverData.offPeakLivesNeededAsCents, &serverData.bardRadius, &serverData.rogueRadius, &serverData.serverName, &serverData.attackSpeedName}, func() error {
			serverDatas[serverData.serverName] = &serverData
			return nil
		}))
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
	http.HandleFunc("/api/online", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	for serverName, serverData := range serverDatas {
		getRowsBlocking(Return12ServerDeaths, func(rows pgx.Rows) {
			var death death
			handleFatalPgx(pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
				serverData.deaths = append(serverData.deaths, death)
				return nil
			}))
		}, serverName)
		getRowsBlocking(Return14ServerEvents, func(rows pgx.Rows) {
			var event event
			handleFatalPgx(pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
				serverData.events = append(serverData.events, event)
				return nil
			}))
		}, serverName)
		getRowsBlocking(Return7ServerFactions, func(rows pgx.Rows) {
			var faction faction
			handleFatalPgx(pgx.ForEachRow(rows, []any{&faction.name, &faction.partyUuid}, func() error {
				serverData.factions = append(serverData.factions, faction)
				return nil
			}))
		}, serverName)
		getRowsBlocking(Return7ServerBandits, func(rows pgx.Rows) {
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
	ID                        string       `json:"id"`
	//		LocationID                string       `json:"location_id"`
	LineItems                 []struct {
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
	}   `json:"line_items"`
	//		Fulfillments              []Fulfillment `json:"fulfillments"`
	CreatedAt                 string       `json:"created_at"`
	UpdatedAt                 string       `json:"updated_at"`
	State                     string       `json:"state"`
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
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AmountMoney money     `json:"amount_money"`
	TipMoney    money     `json:"tip_money"`
	Status      string    `json:"status"`
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
	TotalMoney    money `json:"total_money"`
	ApprovedMoney money `json:"approved_money"`
	//		ReceiptNumber  string         `json:"receipt_number"`
	ReceiptURL     string         `json:"receipt_url"`
	//		DelayAction    string         `json:"delay_action"`
	//		DelayedUntil   time.Time      `json:"delayed_until"`
	//		ApplicationDetails ApplicationDetails `json:"application_details"`
	//		VersionToken   string         `json:"version_token"`
}
var donations = func() []order {
	req := getFatalRequest("https://connect.squareup.com/v2/payments?location_id=" + os.Getenv("SQUARE_LOCATION_ID"), nil)
	addSquareHeaders := func(request *http.Request) {
		request.Header.Add("Authorization", "Bearer " + os.Getenv("SQUARE_ACCESS_TOKEN"))
		request.Header.Add("Content-Type", "application/json")
	}
	addSquareHeaders(req)

	var paymentIds []string
	for _, payment := range getJsonT[struct {Payments []payment `json:"payments"`}](req).Payments {
		paymentIds = append(paymentIds, payment.OrderID)
	}

	requestJson, err := json.Marshal(struct {
		LocationID string   `json:"location_id"`
		OrderIDs   []string `json:"order_ids"`
	} {
		LocationID: os.Getenv("SQUARE_LOCATION_ID"),
		OrderIDs: paymentIds,
		})
	handleFatalErr(err)
	req = getFatalRequest("https://connect.squareup.com/v2/orders/batch-retrieve", bytes.NewBuffer(requestJson))
	addSquareHeaders(req) // shirley this keeps their correct order

	orders := getJsonT[struct {Orders []order `json:"orders"`}](req).Orders
	var orderIds []string
	for _, orderId := range orders {
		orderIds = append(orderIds, orderId.ID)
	}
	getRowsBlocking(ReturnUnsuccessfulTransactions, func(rows pgx.Rows) {
		var missedOrderId string
		handleFatalPgx(pgx.ForEachRow(rows, []any{&missedOrderId}, func() error {
			// TODO -> handle missed transactions
			return nil
		}))
	}, orderIds)

	return orders
}()
func init() {
	http.HandleFunc("/api/donations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET": { // square's payment.create
			payment := getJsonT[struct {
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
					}   `json:"data"`
				} `json:"payload"`
			}](r).Payload.Data.Object.Payment
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
						for _, data := range lineItemDatas {
							if data.gamemodeName + "-" + data.itemName == donationRequest[i].LineItemName {
								if data.isPlural {
									donationRequest[i].LineItemAmount = int(math.Max(float64(donationRequest[i].LineItemAmount), 1))
								} else {
									donationRequest[i].LineItemAmount = 1 // TODO -> maybe allow 0 amount
								}
								donationRequest[i].LineItemCostInCents = data.itemPriceInCents
								continue outer
							}
						}
						// else
						log.Println("err: invalid donationRequest line item")
						return // door nigga from game of thrones
					}

					type LineItem struct {
						Quantity       string `json:"quantity"`
						ItemType       string `json:"item_type"`
						Name           string `json:"name"`
						BasePriceMoney money  `json:"base_price_money"`
					}
					var lineItems []LineItem
					for _, lineItem := range donationRequest {
						lineItems = append(lineItems, LineItem {
							Quantity: strconv.Itoa(lineItem.LineItemAmount),
							ItemType: "ITEM",
							Name: lineItem.LineItemName + "," + lineItem.Username,
							BasePriceMoney: money{
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
						TotalMoney         money `json:"total_money"`
						TaxMoney           money `json:"tax_money"`
						DiscountMoney      money `json:"discount_money"`
						TipMoney           money `json:"tip_money"`
						ServiceChargeMoney money `json:"service_charge_money"`
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
							reqBody, err := json.Marshal(reqData);
							handleNonFatalErr(err)
							req, err := http.NewRequest("POST", "https://connect.squareup.com/v2/online-checkout/payment-links", bytes.NewBuffer(reqBody));
							handleNonFatalErr(err)

							req.Header.Set("Square-Version", "2025-02-20")
							req.Header.Set("Authorization", "Bearer " + os.Getenv("SQUARE_ACCESS_TOKEN"))
							req.Header.Set("Content-Type", "application/json")

							resp, err := (&http.Client{}).Do(req)
							handleNonFatalErr(err)
							defer resp.Body.Close()

							type LineItemResponse struct {
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
								Version                 int   `json:"version"`
								TotalMoney              money `json:"total_money"`
								TotalTaxMoney           money `json:"total_tax_money"`
								TotalDiscountMoney      money `json:"total_discount_money"`
								TotalTipMoney           money `json:"total_tip_money"`
								TotalServiceChargeMoney money `json:"total_service_charge_money"`
								NetAmountDueMoney       money `json:"net_amount_due_money"`
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
							handleNonFatalErr(json.NewDecoder(resp.Body).Decode(&paymentLinkResp))
							_, err = w.Write([]byte(paymentLinkResp.PaymentLink.URL))
							handleNonFatalErr(err)
			}
		}
		}
	})
}
type DiscordMessage struct {
	Type           int           `json:"type"`
	Content        string        `json:"content"`
	Mentions       []interface{} `json:"mentions"`
	MentionRoles   []interface{} `json:"mention_roles"`
	Attachments    []struct {
		ID          string `json:"id"`
		Filename    string `json:"filename"`
		Size        int    `json:"size"`
		URL         string `json:"url"`
		ProxyURL    string `json:"proxy_url"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		ContentType string `json:"content_type"`
	}  `json:"attachments"`
	Embeds         []interface{} `json:"embeds"`
	Timestamp      string        `json:"timestamp"`
	EditedTimestamp interface{}   `json:"edited_timestamp"`
	Flags          int           `json:"flags"`
	Components     []interface{} `json:"components"`
	ID             string        `json:"id"`
	ChannelID      string        `json:"channel_id"`
	Author         struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Avatar        string `json:"avatar"`
		Discriminator string `json:"discriminator"`
		GlobalName    string `json:"global_name"`
	}        `json:"author"`
	Pinned         bool          `json:"pinned"`
	MentionEveryone bool         `json:"mention_everyone"`
	Reactions      []struct {
		Emoji struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"emoji"`
		Count int `json:"count"`
	}    `json:"reactions"`
}
func getDiscordMessages(channelId string) []DiscordMessage {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/channels/" + channelId + "/messages?limit=50", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	return getJsonT[[]DiscordMessage](req)
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
type lineItemData struct {
	gamemodeName string
	itemName string
	itemPriceInCents int
	itemDescription string
	isPlural bool

	itemPriceInDollars int
}
var lineItemDatas = func() []lineItemData {
	// TODO query
	offPeakLivesNeeded := float32(serverDatas["hcf"].offPeakLivesNeededAsCents) / 100.0
	lineItemData := []LineItemData {}
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
	networkPlayers []string
	serverPlayers []string
	newPlayers []newPlayer
	potpissersTips []string
	deaths []death
	messages []string
	events []event
	announcements []DiscordMessage
	changelog []DiscordMessage
	discordMessages []DiscordMessage
	donations []order
	offPeakLivesNeeded float32
	peakLivesNeeded float32
	lineItemData []lineItemData
}
var home = getHome()
var mz = getMz()
var hcf = getHcf()

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
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
//func init() {
//}
func main() {
	var postgresPool *pgxpool.Pool
	{
		var err error
		postgresPool, err = pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
		if err != nil {
			log.Fatal(err)
		}
	}
	defer postgresPool.Close()

	getRowsBlocking := func(query string, bar func(rows pgx.Rows), params ...any) {
		rows, err := postgresPool.Query(context.Background(), query, params...)
		defer rows.Close()
		if err != nil {
			log.Fatal(err)
		}
		bar(rows)
	}

	fetchTips := func(tipsName string) []string {
		var tips []string
		getRowsBlocking(ReturnServerTips, func(rows pgx.Rows) {
			var tipMessage string
			_, err := pgx.ForEachRow(rows, []any{&tipMessage}, func() error {
				tips = append(tips, tipMessage)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, tipsName)
		return tips
	}

	potpissersTips, cubecoreTips, mzTips, cubecoreClassTips := fetchTips("null"), fetchTips("cubecore"), fetchTips("minez"), fetchTips("cubecore_classes")

	type NewPlayer struct {
		PlayerUuid string `json:"playerUuid"`
		Referrer  string `json:"referrer"`
		Timestamp time.Time `json:"timestamp"`
		RowNumber int `json:"rowNumber"`
	}
	var newPlayers []NewPlayer
	{
		getRowsBlocking(Return12NewPlayers, func(rows pgx.Rows) {
			var death NewPlayer
			_, err := pgx.ForEachRow(rows, []any{&death.PlayerUuid, &death.Referrer, &death.Timestamp, &death.RowNumber}, func() error {
				newPlayers = append(newPlayers, death)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		})
	}
	var newPlayersMu sync.RWMutex
	http.HandleFunc("/api/new-players", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[NewPlayer](r, func(newT *NewPlayer, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newT)}, &newPlayersMu, &newPlayers)
	})

	type Death struct {
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
	var deaths []Death
	{
		getRowsBlocking(Return12Deaths, func(rows pgx.Rows) {
			var death Death
			_, err := pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
				deaths = append(deaths, death)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		})
	}
	var deathsMu sync.RWMutex
	http.HandleFunc("/api/deaths", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[Death](r, func(newDeath *Death, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &deathsMu, &deaths)
	})

	type Event struct {
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
	var events []Event
	{
		getRowsBlocking(Return14Events, func(rows pgx.Rows) {
			var event Event
			_, err := pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
				events = append(events, event)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		})
	}
	var eventsMu sync.RWMutex
	http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[Event](r, func(newDeath *Event, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &eventsMu, &events)
	})
	http.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {
	})

	type Faction struct {
		Name string
		PartyUuid string
	}
	type Bandit struct {
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
	type ServerData struct {
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
		Deaths []Death
		Events []Event
//		Transaction []Transaction TODO
		Messages []string
		Videos []string

		Factions []Faction
		Bandits []Bandit
	}
	serverDatas := make(map[string]*ServerData)
	getRowsBlocking(ReturnAllServerData, func(rows pgx.Rows) {
		var serverData ServerData
		_, err := pgx.ForEachRow(rows, []any{&serverData.DeathBanMinutes, &serverData.WorldBorderRadius, &serverData.SharpnessLimit, &serverData.PowerLimit, &serverData.ProtectionLimit, &serverData.RegenLimit, &serverData.StrengthLimit, &serverData.IsWeaknessEnabled, &serverData.IsBardPassiveDebuffingEnabled, &serverData.DtrFreezeTimer, &serverData.DtrMax, &serverData.DtrMaxTime, &serverData.DtrOffPeakFreezeTime, &serverData.OffPeakLivesNeededAsCents, &serverData.BardRadius, &serverData.RogueRadius, &serverData.ServerName, &serverData.AttackSpeedName}, func() error {
			serverDatas[serverData.ServerName] = &serverData
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/api/servers/", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
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
	for serverName, serverData := range serverDatas {
		getRowsBlocking(Return12ServerDeaths, func(rows pgx.Rows) {
			var death Death
			_, err := pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
				serverData.Deaths = append(serverData.Deaths, death)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
		getRowsBlocking(Return14ServerEvents, func(rows pgx.Rows) {
			var event Event
			_, err := pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
				serverData.Events = append(serverData.Events, event)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
		getRowsBlocking(Return7ServerFactions, func(rows pgx.Rows) {
			var faction Faction
			_, err := pgx.ForEachRow(rows, []any{&faction.Name, &faction.PartyUuid}, func() error {
				serverData.Factions = append(serverData.Factions, faction)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
		getRowsBlocking(Return7ServerBandits, func(rows pgx.Rows) {
			var bandit Bandit
			_, err := pgx.ForEachRow(rows, []any{&bandit.UserUuid, &bandit.DeathId, &bandit.Timestamp, &bandit.ExpirationTimestamp, &bandit.DeathMessage, &bandit.DeathWorld, &bandit.DeathX, &bandit.DeathY, &bandit.DeathZ}, func() error {
				serverData.Bandits = append(serverData.Bandits, bandit)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}, serverName)
	}

	type Donation struct {
		ID          string `json:"id"`
		LocationID  string `json:"location_id"`
		CreatedAt   string `json:"created_at"`
		Tenders     []struct {
			ID            string `json:"id"`
			AmountMoney   struct {
				Amount   int    `json:"amount"`
				Currency string `json:"currency"`
			} `json:"amount_money"`
			CreatedAt string `json:"created_at"`
		} `json:"tenders"`
		Refunds     []struct {
			ID            string `json:"id"`
			AmountMoney   struct {
				Amount   int    `json:"amount"`
				Currency string `json:"currency"`
			} `json:"amount_money"`
		} `json:"refunds"`
	}
	type DonationsResponse struct {
		Transactions []Donation `json:"transactions"`
	}
	donations := func() []Donation {
		req, err := http.NewRequest("GET", "https://connect.squareup.com/v2/locations/" + os.Getenv("SQUARE_LOCATION_ID") + "/transactions", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer " + os.Getenv("SQUARE_ACCESS_TOKEN"))
		req.Header.Add("Content-Type", "application/json") // TODO ?

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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
//		println(body) // TODO -> use Decode
		var donationsResponse DonationsResponse
		err = json.Unmarshal(body, &donationsResponse)
		if err != nil {
			log.Fatal(err)
		}
		return donationsResponse.Transactions
	}()
	http.HandleFunc("/api/donations", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

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
	type Message struct {
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
	getDiscordMessages := func(channelId string) []Message {
		req, err := http.NewRequest("GET", "https://discord.com/api/v10/channels/" + channelId + "/messages?limit=50", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
		return getJsonTSlice[Message](req)
	}
	discordAnnouncements, changelog, discordMessages := getDiscordMessages("1265836245678948464"), getDiscordMessages("1346008874830008375"), getDiscordMessages("1245300045188956255")
	// TODO -> store last checked time and then check for every join or something + refresh button + reddit too

	var messages []string // TODO -> make player name clickable
	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		// TODO messages + ServerData.Messages
	})

	type LineItemData struct {
		ServerName string
		ItemName string
		ItemPriceInDollars int
		ItemPriceInCents int
		ItemDescription string
		IsPlural bool
	}
	lineItemDataImmutable := map[string]LineItemData {
		"hcf-life": {
			ServerName: "hcf",
			ItemName: "life",
			ItemPriceInDollars: 4,
			ItemPriceInCents: 400,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: true,
		},
		"hcf-basic": {
			ServerName: "hcf",
			ItemName: "basic",
			ItemPriceInDollars: 8,
			ItemPriceInCents: 800,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
		"hcf-gold": {
			ServerName: "hcf",
			ItemName: "gold",
			ItemPriceInDollars: 16,
			ItemPriceInCents: 1600,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
		"hcf-diamond": {
			ServerName: "hcf",
			ItemName: "diamond",
			ItemPriceInDollars: 24,
			ItemPriceInCents: 2400,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
		"hcf-ruby": {
			ServerName: "hcf",
			ItemName: "ruby",
			ItemPriceInDollars: 32,
			ItemPriceInCents: 3200,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},

		"mz-life": {
			ServerName: "mz",
			ItemName: "life",
			ItemPriceInDollars: 4,
			ItemPriceInCents: 400,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: true,
		},
		"mz-basic": {
			ServerName: "mz",
			ItemName: "basic",
			ItemPriceInDollars: 6,
			ItemPriceInCents: 600,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
		"mz-gold": {
			ServerName: "mz",
			ItemName: "gold",
			ItemPriceInDollars: 12,
			ItemPriceInCents: 1200,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
		"mz-diamond": {
			ServerName: "mz",
			ItemName: "diamond",
			ItemPriceInDollars: 18,
			ItemPriceInCents: 1800,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
		"mz-ruby": {
			ServerName: "mz",
			ItemName: "ruby",
			ItemPriceInDollars: 24,
			ItemPriceInCents: 2400,
			ItemDescription: `/revive (username). removes deathban (alts aren't affected). current revive life cost: {{
.MainTemplateData.OffPeakLivesNeeded }} & {{ .MainTemplateData.PeakLivesNeeded }} during events`, // TODO -> db + ingame
			IsPlural: false,
		},
	}
	http.HandleFunc("/api/donate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
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
			for i := range donationRequest {
//				var foo []LineItemData
//				bar := func() {
//					for j := range foo {
//						if foo[j].ServerName + "-" + foo[j].ItemName == donationRequest[i].LineItemName {
//							if foo[j].IsPlural {
//								donationRequest[i].LineItemAmount = int(math.Max(float64(donationRequest[i].LineItemAmount), 1))
//							} else {
//								donationRequest[i].LineItemAmount = 1 // TODO -> maybe allow 0 amount
//							}
//							donationRequest[i].LineItemCostInCents = foo[j].ItemPriceInCents
//							return
//						}
//					}
//					// else
//					log.Println("error: invalid donationRequest line item")
//					return // door nigga from game of thrones
//				}
				data, exists := lineItemDataImmutable[donationRequest[i].LineItemName]
				if !exists {
					log.Println("error: invalid donationRequest line item")
					return // door nigga from game of thrones
				} else {
					if data.IsPlural {
						donationRequest[i].LineItemAmount = int(math.Max(float64(donationRequest[i].LineItemAmount), 1))
					} else {
						donationRequest[i].LineItemAmount = 1 // TODO -> maybe allow 0 amount
					}
					donationRequest[i].LineItemCostInCents = data.ItemPriceInCents
				}
			}

			type BasePriceMoney struct {
				Amount   int    `json:"amount"`
				Currency string `json:"currency"`
			}
			type LineItem struct {
				Quantity       string `json:"quantity"`
				ItemType       string `json:"item_type"`
				Name           string `json:"name"`
				BasePriceMoney BasePriceMoney `json:"base_price_money"`
			}
			var lineItems []LineItem
			for _, lineItem := range donationRequest {
				lineItems = append(lineItems, LineItem {
					Quantity: strconv.Itoa(lineItem.LineItemAmount),
					ItemType: "ITEM",
					Name: lineItem.LineItemName + "," + lineItem.Username,
					BasePriceMoney: BasePriceMoney {
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
				TotalMoney           BasePriceMoney `json:"total_money"`
				TaxMoney             BasePriceMoney `json:"tax_money"`
				DiscountMoney        BasePriceMoney `json:"discount_money"`
				TipMoney             BasePriceMoney `json:"tip_money"`
				ServiceChargeMoney   BasePriceMoney `json:"service_charge_money"`
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
						RedirectURL: "potpissers.com/donate",
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
						BasePriceMoney BasePriceMoney `json:"base_price_money"`

						UID                      string `json:"uid"`
						VariationTotalPriceMoney BasePriceMoney  `json:"variation_total_price_money"`
						GrossSalesMoney          BasePriceMoney  `json:"gross_sales_money"`
						TotalTaxMoney            BasePriceMoney  `json:"total_tax_money"`
						TotalDiscountMoney       BasePriceMoney  `json:"total_discount_money"`
						TotalMoney               BasePriceMoney  `json:"total_money"`
						TotalServiceChargeMoney  BasePriceMoney  `json:"total_service_charge_money"`
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
						TotalMoney            BasePriceMoney         `json:"total_money"`
						TotalTaxMoney         BasePriceMoney         `json:"total_tax_money"`
						TotalDiscountMoney    BasePriceMoney         `json:"total_discount_money"`
						TotalTipMoney         BasePriceMoney         `json:"total_tip_money"`
						TotalServiceChargeMoney BasePriceMoney       `json:"total_service_charge_money"`
						NetAmountDueMoney     BasePriceMoney         `json:"net_amount_due_money"`
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
	})

	{
		mojangUsernameProxyEndpoint := "/api/proxy/mojang/username/"
		http.HandleFunc(mojangUsernameProxyEndpoint, func(w http.ResponseWriter, r *http.Request) {
			resp, err := http.Get("https://api.minecraftservices.com/minecraft/profile/lookup/name/" + strings.TrimPrefix(r.URL.Path, mojangUsernameProxyEndpoint))
			if err != nil {
				log.Println(err)
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Println(err)
				}
			}(resp.Body)

			// TODO -> headers ?
			w.WriteHeader(resp.StatusCode)

			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Println(err)
			}
		})
	}

	getMainTemplate := func(fileName string) *template.Template {
		mainTemplate, err := template.ParseFiles("main.html", fileName)
		if err != nil {
			log.Fatal(err)
		}
		return mainTemplate
	}
	homeTemplate, mzTemplate, hcfTemplate := getMainTemplate("main-home.html"), getMainTemplate("main-mz.html"), getMainTemplate("main-hcf.html")
	type MainTemplateData struct {
		NetworkPlayers []string
		ServerPlayers []string
		NewPlayers []NewPlayer
		PotpissersTips []string
		Deaths []Death
		Messages []string
		Events []Event
		Announcements []Message
		Changelog []Message
		DiscordMessages []Message
		Donations []Donation
		OffPeakLivesNeeded float32
		PeakLivesNeeded float32
		LineItemData map[string]LineItemData
	}
	getHome := func() []byte {
		var buffer bytes.Buffer
		offPeakLivesNeeded := float32(serverDatas["hcf"].OffPeakLivesNeededAsCents / 100.0)
		err := homeTemplate.Execute(&buffer, struct {
			MainTemplateData MainTemplateData
		}{
			MainTemplateData {
				NetworkPlayers: currentPlayers,
				ServerPlayers: serverDatas["hub"].CurrentPlayers,
				NewPlayers: newPlayers,
				PotpissersTips:     potpissersTips,
				Deaths:             deaths,
				Messages:           messages,
				Events:             events,
				Announcements:      discordAnnouncements,
				Changelog:          changelog,
				DiscordMessages:    discordMessages,
				Donations:          donations,
				OffPeakLivesNeeded: offPeakLivesNeeded,
				PeakLivesNeeded:    offPeakLivesNeeded / 2,
				LineItemData:       lineItemDataImmutable,
				},
		})
		if err != nil {
			log.Fatal(err)
		}
		return buffer.Bytes()
	}
	getMz := func() []byte {
		var buffer bytes.Buffer
		mzData := serverDatas["mz"]
		offPeakLivesNeeded := float32(serverDatas["hcf"].OffPeakLivesNeededAsCents / 100.0)
		err := mzTemplate.Execute(&buffer, struct {
			MainTemplateData MainTemplateData

			AttackSpeed string

			MzTips []string
			Bandits []Bandit
		}{
			MainTemplateData: MainTemplateData {
				NetworkPlayers: currentPlayers,
				ServerPlayers: mzData.CurrentPlayers,
				NewPlayers: newPlayers,
				PotpissersTips: potpissersTips,
				Deaths: mzData.Deaths,
				Messages: mzData.Messages,
				Events: mzData.Events,
				Announcements: discordAnnouncements,
				Changelog: changelog,
				DiscordMessages: discordMessages,
				Donations: donations,
				OffPeakLivesNeeded: offPeakLivesNeeded,
				PeakLivesNeeded: offPeakLivesNeeded / 2,
				LineItemData:       lineItemDataImmutable,
			},

			AttackSpeed: mzData.AttackSpeedName,

			MzTips: mzTips,
			Bandits: mzData.Bandits,
			})
		if err != nil {
			log.Fatal(err)
		}
		return buffer.Bytes()
	}
	getHcf := func() []byte {
		var buffer bytes.Buffer
		serverData := serverDatas["hcf"]
		offPeakLivesNeeded := float32(serverData.OffPeakLivesNeededAsCents / 100.0)
		err := hcfTemplate.Execute(&buffer, struct {
			MainTemplateData MainTemplateData

			AttackSpeed string

			DeathBanMinutes int
			LootFactor int
			BorderSize int

			SharpnessLimit int
			ProtectionLimit int
			PowerLimit int
			RegenLimit int
			StrengthLimit int
			IsWeaknessEnabled bool
			IsBardPassiveDebuffingEnabled bool
			DtrMax float32

			CubecoreTips []string
			ClassTips []string
			Factions []Faction
		}{
			MainTemplateData: MainTemplateData {
				NetworkPlayers: currentPlayers,
				ServerPlayers: serverData.CurrentPlayers,
				NewPlayers: newPlayers,
				PotpissersTips: potpissersTips,
				Deaths: deaths,
				Messages: messages,
				Events: serverData.Events,
				Announcements: discordAnnouncements,
				Changelog: changelog,
				DiscordMessages: discordMessages,
				Donations: donations,
				OffPeakLivesNeeded: offPeakLivesNeeded,
				PeakLivesNeeded: offPeakLivesNeeded / 2,
				LineItemData:       lineItemDataImmutable,
			},

			AttackSpeed: serverData.AttackSpeedName,

			DeathBanMinutes: serverData.DeathBanMinutes,
//			LootFactor: serverDatas["hcf"]., // TODO -> defaultLootFactor
			BorderSize: serverData.WorldBorderRadius,

			SharpnessLimit: serverData.SharpnessLimit,
			ProtectionLimit: serverData.ProtectionLimit,
			PowerLimit: serverData.PowerLimit,
			RegenLimit: serverData.RegenLimit,
			StrengthLimit: serverData.StrengthLimit,
			IsWeaknessEnabled: serverData.IsWeaknessEnabled,
			IsBardPassiveDebuffingEnabled: serverData.IsBardPassiveDebuffingEnabled,
			DtrMax: serverData.DtrMax,

			CubecoreTips: cubecoreTips,
			ClassTips: cubecoreClassTips,
			Factions: serverData.Factions,
			})
		if err != nil {
			log.Fatal(err)
		}
		return buffer.Bytes()
	}
	home, mz, hcf := getHome(), getMz(), getHcf()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(home)
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(mz)
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(hcf)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	err := http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem",  "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}

const ReturnServerTips = `SELECT tip_message
FROM server_tips
         JOIN servers ON server_tips.server_id = servers.id
WHERE name = $1`
const Return12Deaths = `SELECT name,
       victim_user_fight_id,
       timestamp,
       victim_uuid,
       bukkit_victim_inventory,
       death_world,
       death_x,
       death_y,
       death_z,
       death_message,
       killer_uuid,
       bukkit_kill_weapon,
       bukkit_killer_inventory
FROM user_deaths
         JOIN servers ON user_deaths.server_id = servers.id
ORDER BY timestamp DESC
LIMIT 12`
const Return12ServerDeaths = `SELECT name,
       victim_user_fight_id,
       timestamp,
       victim_uuid,
       bukkit_victim_inventory,
       death_world,
       death_x,
       death_y,
       death_z,
       death_message,
       killer_uuid,
       bukkit_kill_weapon,
       bukkit_killer_inventory
FROM user_deaths
         JOIN servers ON user_deaths.server_id = servers.id
WHERE server_id = (SELECT id FROM servers WHERE name = $1)
ORDER BY timestamp DESC
LIMIT 12`
const Return12NewPlayers = `SELECT user_uuid, referrer, timestamp, ROW_NUMBER() OVER (ORDER BY timestamp) AS row_number
FROM user_referrals
ORDER BY timestamp
LIMIT 12`
const Return14Events = `SELECT start_timestamp,
       loot_factor,
       max_timer,
       is_movement_restricted,
       CASE WHEN end_timestamp IS NOT NULL THEN capping_user_uuid END AS capping_user_uuid,
       end_timestamp,
       capping_party_uuid,
       world,
       x,
       y,
       z,
       servers.name                                                   AS server_name,
       arena_data.name                                                AS arena_name,
       creator
FROM koths
         JOIN server_koths ON server_koths_id = server_koths.id
         JOIN servers ON servers.id = server_koths.server_id
         JOIN arena_data ON arena_data.id = server_koths.arena_id
ORDER BY end_timestamp IS NULL, end_timestamp
LIMIT 14`
const Return14ServerEvents = `SELECT start_timestamp,
       loot_factor,
       max_timer,
       is_movement_restricted,
       CASE WHEN end_timestamp IS NOT NULL THEN capping_user_uuid END AS capping_user_uuid,
       end_timestamp,
       capping_party_uuid,
       world,
       x,
       y,
       z,
       servers.name                                                   AS server_name,
       arena_data.name                                                AS arena_name,
       creator
FROM koths
         JOIN server_koths ON server_koths_id = server_koths.id
         JOIN servers ON servers.id = server_koths.server_id
         JOIN arena_data ON arena_data.id = server_koths.arena_id
WHERE server_id = (SELECT id FROM servers WHERE name = $1)
ORDER BY end_timestamp IS NULL, end_timestamp
LIMIT 14`
const ReturnAllServerData = `SELECT death_ban_minutes,
       world_border_radius,
       sharpness_limit,
       power_limit,
       protection_limit,
       bard_regen_level,
       bard_strength_level,
       is_weakness_enabled,
       is_bard_passive_debuffing_enabled,
       dtr_freeze_timer,
       dtr_max,
       dtr_max_time,
       dtr_off_peak_freeze_time,
       off_peak_lives_needed_as_cents,
       bard_radius,
       rogue_radius,
       servers.name,
       attack_speeds.attack_speed_name
FROM server_data
         JOIN servers ON id = server_id
         JOIN attack_speeds ON attack_speed_id = attack_speeds.id`
const ReturnAllOnlinePlayers = `SELECT user_name, name
FROM online_players
         JOIN servers ON server_id = servers.id`
const Return7ServerFactions = `SELECT name, party_uuid
FROM factions
WHERE server_id = (SELECT id FROM servers WHERE name = $1)
LIMIT 7`
const Return7ServerBandits = `SELECT user_uuid,
       death_id,
       timestamp,
       expiration_timestamp,
       death_message,
       death_world,
       death_x,
       death_y,
       death_z
FROM bandits
         JOIN user_deaths on bandits.death_id = user_deaths.id
WHERE bandits.server_id = (SELECT id FROM servers WHERE name = $1)
  AND expiration_timestamp > NOW()
ORDER BY timestamp DESC
LIMIT 7`

func handlePutJson[T any](r *http.Request, decodeJson func(*T, *http.Request) error, mutex *sync.RWMutex, collection *[]T) {
	if r.Method == "PATCH" {
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			var newT T
			err := decodeJson(&newT, r)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Lock()
			*collection = append([]T{newT}, *collection...) // TODO -> this is necessary because html/css and go's templating can't handle reversing it for some reason. go's templater could maybe do it but it seems like more processing than this takes
			mutex.Unlock()
		}
	}
}
func getJsonTSlice[T any](request *http.Request) []T {
	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)
	var messages []T
	err = json.NewDecoder(resp.Body).Decode(&messages)
	if err != nil {
		log.Fatal(err)
	}
	return messages
}
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var potpissersTips []string
var cubecoreTips []string
var cubecoreClassTips []string
var mzTips []string

func init() { // TODO -> move this to getting tips by name
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
	PlayerUuid string    `json:"player_uuid"`
	Referrer   *string   `json:"referrer"`
	Timestamp  time.Time `json:"timestamp"`
	RowNumber  int       `json:"row_number"`

	PlayerName string `json:"player_name"`
}

var newPlayers = func() []newPlayer {
	var newPlayers []newPlayer
	getRowsBlocking("SELECT * FROM get_10_newest_players()", func(rows pgx.Rows) {
		var death newPlayer
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.PlayerUuid, &death.Referrer, &death.Timestamp, &death.RowNumber}, func() error {
			newPlayers = append(newPlayers, death)
			return nil
		}))
	})

	for i := range newPlayers {
		resp, err := http.Get("https://api.minecraftservices.com/minecraft/profile/lookup/" + newPlayers[i].PlayerUuid)
		if err != nil {
			log.Fatal(err)
		}
		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		newPlayers[i].PlayerName = result["name"].(string)
	}

	println("new players done")
	return newPlayers
}()

type death struct {
	ServerName        string      `json:"server_name"`
	VictimUserFightId pgtype.Int4 `json:"victim_user_fight_id"`
	Timestamp         time.Time   `json:"timestamp"`
	VictimUuid        string      `json:"victim_uuid"`
	// TODO victim inventory
	DeathWorldName string  `json:"death_world_name"`
	DeathX         int     `json:"death_x"`
	DeathY         int     `json:"death_y"`
	DeathZ         int     `json:"death_z"`
	DeathMessage   string  `json:"death_message"`
	KillerUuid     *string `json:"killer_uuid"`
	// TODO killer weapon
	// TODO killer inventory
}

var deaths = func() []death {
	var deaths []death
	getRowsBlocking("SELECT * FROM get_12_latest_network_deaths()", func(rows pgx.Rows) {
		var death death
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
			deaths = append(deaths, death)
			return nil
		}))
	})
	println("deaths done")
	return deaths
}()

type event struct {
	ServerKothsId        int       `json:"server_koths_id"`
	StartTimestamp       time.Time `json:"start_timestamp"`
	LootFactor           int       `json:"loot_factor"`
	MaxTimer             int       `json:"max_timer"`
	IsMovementRestricted bool      `json:"is_movement_restricted"`
	CappingUserUUID      *string   `json:"capping_user_uuid"`
	EndTimestamp         time.Time `json:"end_timestamp"`
	CappingPartyUUID     *string   `json:"capping_party_uuid"`
	CapMessage           *string   `json:"cap_message"`
	World                string    `json:"world"`
	X                    int       `json:"x"`
	Y                    int       `json:"y"`
	Z                    int       `json:"z"`
	ServerName           string    `json:"server_name"`
	ArenaName            string    `json:"arena_name"`
	Creator              string    `json:"creator"`
}

var events = func() []event {
	var events []event
	getRowsBlocking("SELECT * FROM get_14_newest_network_koths()", func(rows pgx.Rows) {
		var event event
		handleFatalPgx(pgx.ForEachRow(rows, []any{&event.ServerKothsId, &event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
			events = append(events, event)
			return nil
		}))
	})
	println("events done")
	return events
}()

type faction struct {
	name                    string
	partyUuid               string
	frozenUntil             time.Time
	currentMaxDtr           float32
	currentRegenAdjustedDtr float32
}
type bandit struct {
	UserUuid            string    `json:"user_uuid"`
	DeathId             int       `json:"death_id"`
	DeathTimestamp      time.Time `json:"death_timestamp"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp"`
	BanditMessage       string    `json:"bandit_message"`
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
	gamemodeName                  string
	attackSpeedName               string

	currentPlayers []onlinePlayer
	deaths         []death
	events         []event
	donations      []order // TODO impl
	messages       []ingameMessage
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

	println("serverdatas done")
	return serverDatas
}()

type onlinePlayer struct {
	Uuid          string    `json:"uuid"`
	Name          string    `json:"name"`
	ServerName    string    `json:"server_name"`
	ActiveFaction *string   `json:"active_faction"`
	NetworkJoin   time.Time `json:"network_join"`
	ServerJoin    time.Time `json:"server_join"`
}

var currentPlayers = func() []onlinePlayer {
	var currentPlayers []onlinePlayer
	getRowsBlocking("SELECT * FROM get_online_players()", func(rows pgx.Rows) {
		var t onlinePlayer
		handleFatalPgx(pgx.ForEachRow(rows, []any{&t.Uuid, &t.Name, &t.ServerName, &t.ActiveFaction, &t.NetworkJoin, &t.ServerJoin}, func() error {
			currentPlayers = append(currentPlayers, t) // TODO sort names
			serverData := serverDatas[t.ServerName]
			serverData.currentPlayers = append(serverData.currentPlayers, t)
			return nil
		}))
	})
	return currentPlayers
}()

func init() {
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
			handleFatalPgx(pgx.ForEachRow(rows, []any{&event.ServerKothsId, &event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
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
			handleFatalPgx(pgx.ForEachRow(rows, []any{&bandit.UserUuid, &bandit.DeathId, &bandit.DeathTimestamp, &bandit.ExpirationTimestamp, &bandit.BanditMessage}, func() error {
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

var donationsMu sync.RWMutex
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
			home = getHome()
			mz = getMz()
			hcf = getHcf()
			handleSseData(jsonBytes, homeConnections, mzConnections, hcfConnections)
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

func getDiscordMessages(channelId string, apiUrlModifier string) []discordMessage {
	for {
		req, err := http.NewRequest("GET", "https://discord.com/api/v10/channels/"+channelId+"/messages?"+apiUrlModifier+"limit=6", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Bot "+os.Getenv("DISCORD_BOT_TOKEN"))

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var messages []discordMessage
		err = json.Unmarshal(body, &messages)
		if err != nil {
			var response map[string]any
			err = json.Unmarshal(body, &response)
			if err != nil {
				log.Fatal(err)
			}
			retryAfter, ok := response["retry_after"].(float64)
			if !ok {
				log.Fatal(err)
			}
			time.Sleep(time.Duration(retryAfter * float64(time.Second)))
			continue
		}
		return messages
	}
}

const discordGeneralChannelId = "1245300045188956255"

var discordMessages = getDiscordMessages(discordGeneralChannelId, "")
var mostRecentDiscordGeneralMessageId = discordMessages[0].ID
var discordGeneralChan = make(chan struct{}, 1)

const discordChangelogChannelId = "1346008874830008375"

var changelog = getDiscordMessages(discordChangelogChannelId, "")
var mostRecentDiscordChangelogMessageId = "" //changelog[0].ID
var discordChangelogChan = make(chan struct{}, 1)

const discordAnnouncementsChannelId = "1265836245678948464"

var announcements = getDiscordMessages(discordAnnouncementsChannelId, "")
var mostRecentDiscordAnnouncementsMessageId = announcements[0].ID
var discordAnnouncementsChan = make(chan struct{}, 1)

func init() {
	http.HandleFunc("/api/discord/general", func(w http.ResponseWriter, r *http.Request) {
		handleDiscordMessagesUpdate(discordGeneralChan, discordGeneralChannelId, &mostRecentDiscordGeneralMessageId, &discordMessages, "general")
	})
	//	http.HandleFunc("/api/discord/changelog", func(w http.ResponseWriter, r *http.Request) {
	//		handleDiscordMessagesUpdate(discordChangelogChan, discordChangelogChannelId, &mostRecentDiscordChangelogMessageId, &changelog, "changelog")
	//	})
	http.HandleFunc("/api/discord/announcements", func(w http.ResponseWriter, r *http.Request) {
		handleDiscordMessagesUpdate(discordAnnouncementsChan, discordAnnouncementsChannelId, &mostRecentDiscordAnnouncementsMessageId, &announcements, "announcements")
	})
	println("discord done")
}

type ingameMessage struct {
	Uuid       string `json:"uuid"`
	Message    string `json:"message"`
	ServerName string `json:"server_name"`
}

var messages []ingameMessage

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

type redditVideoPost struct {
	YoutubeEmbedUrl string `json:"youtube_embed_url"`
	VideoUrl        string `json:"video_url"`
	PostUrl         string `json:"post_url"`
	Title           string `json:"title_url"`
}

var redditVideoPosts []redditVideoPost

type redditImagePost struct {
	ImageUrl string `json:"image_url"`
	PostUrl  string `json:"post_url"`
}

var redditImagePosts []redditImagePost
var imageRegex = regexp.MustCompile(`(?i)^(https?://)?(i\.redd\.it|i\.imgur\.com)/.*\.(png|jpg|jpeg)$`)
var youtubeVideoIdRegex = regexp.MustCompile(`[?&]v=([a-zA-Z0-9_-]{11})`)
var lastCheckedRedditPostId string
var lastCheckedRedditPostCreatedUtc float64
var redditPostsChannel = make(chan struct{}, 1)

func init() {
	redditVideoPosts, redditImagePosts = getRedditPostData(potpissersRedditApiUrl)
	http.HandleFunc("/api/reddit", func(w http.ResponseWriter, r *http.Request) {
		handleRedditPostDataUpdate()
	})
	println("reddit done")
}

const frontendDirName = "./go-frontend"

func getMainTemplate(fileName string) *template.Template {
	mainTemplate, err := template.ParseFiles(frontendDirName+"/main.html", fileName)
	handleFatalErr(err)
	return mainTemplate
}

var homeTemplate = getMainTemplate(frontendDirName + "/main-home.html")
var mzTemplate = getMainTemplate(frontendDirName + "/main-mz.html")
var hcfTemplate = getMainTemplate(frontendDirName + "/main-hcf.html")

type mainTemplateData struct {
	GamemodeName       string
	BackgroundImageUrl redditImagePost
	NetworkPlayers     []onlinePlayer
	ServerPlayers      []onlinePlayer
	NewPlayers         []newPlayer
	PotpissersTips     []string
	Deaths             []death
	Messages           []ingameMessage
	Events             []event
	Announcements      []discordMessage
	Changelog          []discordMessage
	DiscordMessages    []discordMessage
	Donations          []order
	OffPeakLivesNeeded float32
	PeakLivesNeeded    float32
	LineItemData       []lineItemData
	RedditVideos       []redditVideoPost
	DiscordId          string
}

func getRandomImagePost() redditImagePost {
	return redditImagePosts[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(redditImagePosts))]
}

const discordServerId = "1245300045188956252"

func getHome() []byte {
	var buffer bytes.Buffer
	offPeakLivesNeeded := float32(serverDatas[currentHcfServerName].offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(homeTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData
	}{
		MainTemplateData: mainTemplateData{
			GamemodeName:       "hub",
			BackgroundImageUrl: getRandomImagePost(),
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
			RedditVideos:       redditVideoPosts,
			DiscordId:          discordServerId,
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
			GamemodeName:       "mz",
			BackgroundImageUrl: getRandomImagePost(),
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
			RedditVideos:       redditVideoPosts,
			DiscordId:          discordServerId,
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
			GamemodeName:       "hcf",
			BackgroundImageUrl: getRandomImagePost(),
			NetworkPlayers:     currentPlayers,
			ServerPlayers:      serverData.currentPlayers,
			NewPlayers:         newPlayers,
			PotpissersTips:     potpissersTips,
			Deaths:             serverData.deaths,
			Messages:           serverData.messages,
			Events:             serverData.events,
			Announcements:      announcements,
			Changelog:          changelog,
			DiscordMessages:    discordMessages,
			Donations:          donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded:    offPeakLivesNeeded / 2,
			LineItemData:       lineItemDatas,
			RedditVideos:       redditVideoPosts,
			DiscordId:          discordServerId,
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

var home []byte
var mz []byte
var hcf []byte

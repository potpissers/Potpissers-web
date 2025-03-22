package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
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
	PlayerUuid string    `json:"playerUuid"`
	Referrer   *string   `json:"referrer"`
	Timestamp  time.Time `json:"timestamp"`
	RowNumber  int       `json:"rowNumber"`

	PlayerName string `json:"playerName"`
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

	return newPlayers
}()

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

	return serverDatas
}()

type onlinePlayer struct {
	Name          string
	ServerName    string
	ActiveFaction string
}

var currentPlayers = func() []onlinePlayer {
	var currentPlayers []onlinePlayer
	getRowsBlocking("SELECT * FROM get_online_players()", func(rows pgx.Rows) {
		var t onlinePlayer
		handleFatalPgx(pgx.ForEachRow(rows, []any{&t.Name, &t.ServerName, &t.ActiveFaction}, func() error {
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
}()

type ingameMessage struct {
	Name       string
	Uuid       string
	Message    string
	ServerName string
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
	ThumbnailUrl string
	PostUrl      string
	Title        string
}

var redditVideoPosts []redditVideoPost

type redditImagePost struct {
	ImageUrl string
	PostUrl  string
}

var redditImagePosts []redditImagePost
var imageRegex = regexp.MustCompile(`(?i)^(https?://)?(i\.redd\.it|i\.imgur\.com)/.*\.(png|jpg|jpeg)$`)
var youtubeVideoIdRegex = regexp.MustCompile(`/[?&]v=([a-zA-Z0-9_-]{11})/`)
var redditPostIdRegex = regexp.MustCompile(`/comments/([^/]+)/`)
var lastCheckedRedditPostId string
var redditPostsChannel = make(chan struct{}, 1)

func init() {
	getRedditPostData := func(redditApiUrl string) ([]redditVideoPost, []redditImagePost, string) {
		data := url.Values{}
		data.Set("grant_type", "client_credentials")
		req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("REDDIT_CLIENT_ID")+":"+os.Getenv("REDDIT_CLIENT_SECRET"))))
		client := &http.Client{}
		authResp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer authResp.Body.Close()
		var result map[string]any
		if err := json.NewDecoder(authResp.Body).Decode(&result); err != nil {
			log.Fatal(err)
		}
		redditAccessToken := result["access_token"].(string)

		req, err = http.NewRequest("GET", redditApiUrl, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+redditAccessToken)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		responseJson := getFatalJsonT[struct {
			Kind string `json:"kind"`
			Data struct {
				After     *string `json:"after"`
				Dist      int     `json:"dist"`
				Modhash   string  `json:"modhash"`
				GeoFilter string  `json:"geo_filter"`
				Children  []struct {
					Kind string `json:"kind"`
					Data struct {
						Subreddit   string  `json:"subreddit"`
						Title       string  `json:"title"`
						Selftext    string  `json:"selftext"`
						Author      string  `json:"author"`
						UpvoteRatio float64 `json:"upvote_ratio"`
						Thumbnail   string  `json:"thumbnail"`
						URL         string  `json:"url"`
						NumComments int     `json:"num_comments"`
						Permalink   string  `json:"permalink"`
						CreatedUTC  float64 `json:"created_utc"`
						IsVideo     bool    `json:"is_video"`
						Media       *struct {
							RedditVideo *struct {
								FallbackURL  string `json:"fallback_url"`
								Height       int    `json:"height"`
								Width        int    `json:"width"`
								Duration     int    `json:"duration"`
								ThumbnailURL string `json:"thumbnail_url"`
							} `json:"reddit_video"`
						} `json:"media,omitempty"`
					} `json:"data"`
				} `json:"children"`
			} `json:"data"`
		}](resp)

		var videoPosts []redditVideoPost
		var imagePosts []redditImagePost
		children := responseJson.Data.Children
		for _, child := range children {
			getRedditPostUrl := func(permalink string) string {
				return "https://www.reddit.com" + permalink
			}

			data := child.Data
			linkPostUrl := data.URL

			if imageRegex.MatchString(linkPostUrl) {
				imagePosts = append(imagePosts, redditImagePost{linkPostUrl, getRedditPostUrl(data.Permalink)})
			} else if strings.HasPrefix(linkPostUrl, "https://youtube.com") || strings.HasPrefix(linkPostUrl, "https://youtu.be") {
				videoPosts = append(videoPosts, redditVideoPost{
					ThumbnailUrl: "https://img.youtube.com/vi/" + youtubeVideoIdRegex.FindStringSubmatch(linkPostUrl)[1] + "/hqdefault.jpg",
					PostUrl:      getRedditPostUrl(data.Permalink),
					Title:        data.Title,
				})
			} else if data.Media != nil {
				videoPosts = append(videoPosts, redditVideoPost{
					ThumbnailUrl: data.Media.RedditVideo.ThumbnailURL,
					PostUrl:      getRedditPostUrl(data.Permalink),
					Title:        data.Title,
				})
			}
		}
		return videoPosts, imagePosts, redditPostIdRegex.FindStringSubmatch(children[0].Data.Permalink)[1]
	}
	const potpissersRedditApiUrl = "https://oauth.reddit.com/r/potpissers/new.json?limit=100"
	redditVideoPosts, redditImagePosts, lastCheckedRedditPostId = getRedditPostData(potpissersRedditApiUrl)
	http.HandleFunc("/api/reddit", func(w http.ResponseWriter, r *http.Request) {
		select {
		case redditPostsChannel <- struct{}{}: {
			var newVideoPosts []redditVideoPost
			var newImagePosts []redditImagePost
			newVideoPosts, newImagePosts, lastCheckedRedditPostId = getRedditPostData(potpissersRedditApiUrl + "&after=" + lastCheckedRedditPostId)
			<-redditPostsChannel
		}
		default: return
		}
	})
}

func getMainTemplate(fileName string) *template.Template {
	mainTemplate, err := template.ParseFiles("main.html", fileName)
	handleFatalErr(err)
	return mainTemplate
}

var homeTemplate = getMainTemplate("main-home.html")
var mzTemplate = getMainTemplate("main-mz.html")
var hcfTemplate = getMainTemplate("main-hcf.html")

type mainTemplateData struct {
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
}

func getRandomImagePost() redditImagePost {
	return redditImagePosts[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(redditImagePosts))]
}

func getHome() []byte {
	var buffer bytes.Buffer
	offPeakLivesNeeded := float32(serverDatas[currentHcfServerName].offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(homeTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData
	}{
		MainTemplateData: mainTemplateData{
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
			BackgroundImageUrl: getRandomImagePost(),
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

var home []byte
var mz []byte
var hcf []byte

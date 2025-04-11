package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

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

func handleDiscordMessagesUpdate(channel chan struct{}, discordChannelId string, mostRecentMessageId *string, slice *[]discordMessage, sseMessageType string) {
	select {
	case channel <- struct{}{}:
		{
			newMessages := getDiscordMessages(discordChannelId, "after="+*mostRecentMessageId+"&")
			if len(newMessages) > 0 {
				*mostRecentMessageId = newMessages[0].ID

				home = getMainTemplateBytes("hub")
				hcf = getMainTemplateBytes("hcf")
				mz = getMainTemplateBytes("mz")

				for _, msg := range newMessages {
					*slice = append([]discordMessage{msg}, *slice...)

					jsonData, err := json.Marshal(sseMessage{sseMessageType, msg})
					if err != nil {
						log.Fatal(err)
					}
					handleSseData(jsonData, mainConnections)
				}
			}
			time.Sleep(time.Second)
			<-channel
		}
	default:
		return
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func getRowsBlocking(query string, bar func(rows pgx.Rows), params ...any) {
	rows, err := postgresPool.Query(context.Background(), query, params...)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	bar(rows)
}

func getHome() []byte {
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
			Announcements:      announcements,
			Changelog:          changelog,
			DiscordMessages:    discordMessages,
			Donations:          donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded:    offPeakLivesNeeded / 2,
			LineItemData:       lineItemData,
			},
			})
	if err != nil {
		log.Fatal(err)
	}
	return buffer.Bytes()
}
func getMz() []byte {
	var buffer bytes.Buffer
	mzData := serverDatas["mz"]
	offPeakLivesNeeded := float32(serverDatas["hcf"].OffPeakLivesNeededAsCents / 100.0)
	err := mzTemplate.Execute(&buffer, struct {
		MainTemplateData MainTemplateData

		AttackSpeed string

		MzTips []string
		Bandits []bandit
	}{
		MainTemplateData: MainTemplateData {
			NetworkPlayers: currentPlayers,
			ServerPlayers: mzData.CurrentPlayers,
			NewPlayers: newPlayers,
			PotpissersTips: potpissersTips,
			Deaths: mzData.Deaths,
			Messages: mzData.Messages,
			Events: mzData.Events,
			Announcements: announcements,
			Changelog: changelog,
			DiscordMessages: discordMessages,
			Donations: donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded: offPeakLivesNeeded / 2,
			LineItemData:       lineItemData,
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
func getHcf() []byte {
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
		Factions []faction
	}{
		MainTemplateData: MainTemplateData {
			NetworkPlayers: currentPlayers,
			ServerPlayers: serverData.CurrentPlayers,
			NewPlayers: newPlayers,
			PotpissersTips: potpissersTips,
			Deaths: deaths,
			Messages: messages,
			Events: serverData.Events,
			Announcements: announcements,
			Changelog: changelog,
			DiscordMessages: discordMessages,
			Donations: donations,
			OffPeakLivesNeeded: offPeakLivesNeeded,
			PeakLivesNeeded: offPeakLivesNeeded / 2,
			LineItemData:       lineItemData,
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
func getMainTemplate(fileName string) *template.Template {
	mainTemplate, err := template.ParseFiles("main.html", fileName)
	if err != nil {
		log.Fatal(err)
	}
	return mainTemplate
}

func getTipsBlocking(tipsName string) []string {
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

func getFatalRequest(url string, body io.Reader) *http.Request {
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		log.Fatal(err)
	}
	return req
}

func getDiscordMessages(channelId string) []DiscordMessage {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/channels/" + channelId + "/messages?limit=50", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	return getJsonT[[]DiscordMessage](req)
}

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
func getJsonT[T any](request *http.Request) T {
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
	var messages any
	err = json.NewDecoder(resp.Body).Decode(&messages)
	if err != nil {
		log.Fatal(err)
	}
	return messages
}
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

func handleFatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func getFatalRequest(url string, body io.Reader) *http.Request {
	req, err := http.NewRequest("GET", url, body)
	handleFatalErr(err)
	return req
}
func handleFatalPgx(_ pgconn.CommandTag, err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getRowsBlocking(query string, bar func(rows pgx.Rows), params ...any) {
	rows, err := postgresPool.Query(context.Background(), query, params...)
	defer rows.Close()
	handleFatalErr(err)
	bar(rows)
}

func getHome() []byte {
	var buffer bytes.Buffer
	offPeakLivesNeeded := float32(serverDatas["hcf"].offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(homeTemplate.Execute(&buffer, struct {
		mainTemplateData mainTemplateData
	}{
		mainTemplateData {
			networkPlayers: currentPlayers,
			serverPlayers: serverDatas["hub"].currentPlayers,
			newPlayers: newPlayers,
			potpissersTips:     potpissersTips,
			deaths:             deaths,
			messages:           messages,
			events:             events,
			announcements:      announcements,
			changelog:          changelog,
			discordMessages:    discordMessages,
			donations:          donations,
			offPeakLivesNeeded: offPeakLivesNeeded,
			peakLivesNeeded:    offPeakLivesNeeded / 2,
			lineItemData:       lineItemDatas,
			},
			}))
	return buffer.Bytes()
}
func getMz() []byte {
	var buffer bytes.Buffer
	mzData := serverDatas["mz"]
	offPeakLivesNeeded := float32(serverDatas["hcf"].offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(mzTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData

		AttackSpeed string

		MzTips []string
		Bandits []bandit
	}{
		MainTemplateData: mainTemplateData {
			networkPlayers: currentPlayers,
			serverPlayers: mzData.currentPlayers,
			newPlayers: newPlayers,
			potpissersTips: potpissersTips,
			deaths: mzData.deaths,
			messages: mzData.messages,
			events: mzData.events,
			announcements: announcements,
			changelog: changelog,
			discordMessages: discordMessages,
			donations: donations,
			offPeakLivesNeeded: offPeakLivesNeeded,
			peakLivesNeeded: offPeakLivesNeeded / 2,
			lineItemData:       lineItemDatas,
			},

			AttackSpeed: mzData.attackSpeedName,

			MzTips: mzTips,
			Bandits: mzData.bandits,
			}))
	return buffer.Bytes()
}
func getHcf() []byte {
	var buffer bytes.Buffer
	serverData := serverDatas["hcf"]
	offPeakLivesNeeded := float32(serverData.offPeakLivesNeededAsCents / 100.0)
	handleFatalErr(hcfTemplate.Execute(&buffer, struct {
		MainTemplateData mainTemplateData

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
		MainTemplateData: mainTemplateData {
			networkPlayers: currentPlayers,
			serverPlayers: serverData.currentPlayers,
			newPlayers: newPlayers,
			potpissersTips: potpissersTips,
			deaths: deaths,
			messages: messages,
			events: serverData.events,
			announcements: announcements,
			changelog: changelog,
			discordMessages: discordMessages,
			donations: donations,
			offPeakLivesNeeded: offPeakLivesNeeded,
			peakLivesNeeded: offPeakLivesNeeded / 2,
			lineItemData:       lineItemDatas,
			},

			AttackSpeed: serverData.attackSpeedName,

			DeathBanMinutes: serverData.deathBanMinutes,
			//			LootFactor: serverDatas["hcf"]., // TODO -> defaultLootFactor
			BorderSize: serverData.worldBorderRadius,

			SharpnessLimit: serverData.sharpnessLimit,
			ProtectionLimit: serverData.protectionLimit,
			PowerLimit: serverData.powerLimit,
			RegenLimit: serverData.regenLimit,
			StrengthLimit: serverData.strengthLimit,
			IsWeaknessEnabled: serverData.isWeaknessEnabled,
			IsBardPassiveDebuffingEnabled: serverData.isBardPassiveDebuffingEnabled,
			DtrMax: serverData.dtrMax,

			CubecoreTips: cubecoreTips,
			ClassTips: cubecoreClassTips,
			Factions: serverData.factions,
			}))
	return buffer.Bytes()
}

func handleLocalhostPutJson[T any](r *http.Request, decodeJson func(*T, *http.Request) error, mutex *sync.RWMutex, collection *[]T) {
	if r.Method == "PATCH" {
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			var newT T
			handleFatalErr(decodeJson(&newT, r))
			
			mutex.Lock()
			*collection = append([]T{newT}, *collection...) // TODO -> this is necessary because html/css and go's templating can't handle reversing it for some reason. go's templater could maybe do it but it seems like more processing than this takes
			mutex.Unlock()
		}
	}
}
func getJsonT[T any](request *http.Request) T {
	resp, err := (&http.Client{}).Do(request)
	handleFatalErr(err)
	defer resp.Body.Close()
	var messages any
	handleFatalErr(json.NewDecoder(resp.Body).Decode(&messages))
	return messages
}
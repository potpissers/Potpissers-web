package main

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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
	defaultKothLootFactor         int
	sharpnessLimit                int
	powerLimit                    int
	protectionLimit               int
	regenLimit                    int
	strengthLimit                 int
	isWeaknessEnabled             bool
	isBardPassiveDebuffingEnabled bool
	dtrFreezeTimer                int
	dtrMax                        float32
	offPeakLivesNeededAsCents     int
	timestamp                     time.Time
	serverName                    string
	gameModeName                  string
	attackSpeedName               string
	isInitiallyWhitelisted        bool

	currentPlayers []onlinePlayer
	deaths         []death
	events         []event
	// donations      []order // TODO impl
	messages []ingameMessage
	// videos         []string

	factions []faction
	bandits  []bandit
}

var currentHcfServerName string
var serverDatas = func() map[string]*serverData {
	serverDatas := make(map[string]*serverData)
	getRowsBlocking("SELECT * FROM get_server_datas()", func(rows pgx.Rows) {
		var serverDataBuffer serverData
		handleFatalPgx(pgx.ForEachRow(rows, []any{&serverDataBuffer.deathBanMinutes, &serverDataBuffer.worldBorderRadius, &serverDataBuffer.defaultKothLootFactor, &serverDataBuffer.sharpnessLimit, &serverDataBuffer.powerLimit, &serverDataBuffer.protectionLimit, &serverDataBuffer.regenLimit, &serverDataBuffer.strengthLimit, &serverDataBuffer.isWeaknessEnabled, &serverDataBuffer.isBardPassiveDebuffingEnabled, &serverDataBuffer.dtrFreezeTimer, &serverDataBuffer.dtrMax, &serverDataBuffer.offPeakLivesNeededAsCents, &serverDataBuffer.timestamp, &serverDataBuffer.serverName, &serverDataBuffer.gameModeName, &serverDataBuffer.attackSpeedName, &serverDataBuffer.isInitiallyWhitelisted}, func() error {
			serverData := serverDataBuffer
			serverDatas[serverDataBuffer.gameModeName+serverDataBuffer.serverName] = &serverData
			return nil
		}))
	})
	var currentPotentialHcfServerTimestamp time.Time
	for _, data := range serverDatas {
		if strings.Contains(data.gameModeName, "hcf") && data.timestamp.After(currentPotentialHcfServerTimestamp) && (currentPotentialHcfServerTimestamp.IsZero() || !data.isInitiallyWhitelisted) {
			currentHcfServerName = data.serverName
			currentPotentialHcfServerTimestamp = data.timestamp
		}
	}

	println("serverdatas done")
	return serverDatas
}()

func init() {
	for _, serverData := range serverDatas {
		getRowsBlocking("SELECT * FROM get_12_latest_server_deaths($1, $2)", func(rows pgx.Rows) {
			var death death
			handleFatalPgx(pgx.ForEachRow(rows, []any{&death.GameModeName, &death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
				serverData.deaths = append(serverData.deaths, death)
				return nil
			}))
		}, serverData.gameModeName, serverData.serverName)
		getRowsBlocking("SELECT * FROM get_14_newest_server_koths($1, $2)", func(rows pgx.Rows) {
			var event event
			handleFatalPgx(pgx.ForEachRow(rows, []any{&event.ServerKothsId, &event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.GameModeName, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
				serverData.events = append(serverData.events, event)
				return nil
			}))
		}, serverData.gameModeName, serverData.serverName)
		getRowsBlocking("SELECT * FROM get_7_factions($1, $2)", func(rows pgx.Rows) {
			var faction faction
			handleFatalPgx(pgx.ForEachRow(rows, []any{&faction.name, &faction.partyUuid, &faction.frozenUntil, &faction.currentMaxDtr, &faction.currentRegenAdjustedDtr}, func() error {
				serverData.factions = append(serverData.factions, faction)
				return nil
			}))
		}, serverData.gameModeName, serverData.serverName)
		getRowsBlocking("SELECT * FROM get_7_newest_bandits($1, $2)", func(rows pgx.Rows) {
			var bandit bandit
			handleFatalPgx(pgx.ForEachRow(rows, []any{&bandit.UserUuid, &bandit.DeathId, &bandit.DeathTimestamp, &bandit.ExpirationTimestamp, &bandit.BanditMessage}, func() error {
				serverData.bandits = append(serverData.bandits, bandit)
				return nil
			}))
		}, serverData.gameModeName, serverData.serverName)
	}

	println("serverdatas data done")
}

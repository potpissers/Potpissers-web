package main

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type faction struct {
	Name                    string    `json:"name"`
	PartyUuid               string    `json:"party_uuid"`
	FrozenUntil             time.Time `json:"frozen_until"`
	CurrentMaxDtr           float32   `json:"current_max_dtr"`
	CurrentRegenAdjustedDtr float32   `json:"current_regen_adjusted_dtr"`
}
type bandit struct {
	UserUuid            string    `json:"user_uuid"`
	DeathId             int       `json:"death_id"`
	DeathTimestamp      time.Time `json:"death_timestamp"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp"`
	BanditMessage       string    `json:"bandit_message"`
}
type serverData struct {
	DeathBanMinutes               int       `json:"death_ban_minutes"`
	WorldBorderRadius             int       `json:"world_border_radius"`
	DefaultKothLootFactor         int       `json:"default_koth_loot_factor"`
	SharpnessLimit                int       `json:"sharpness_limit"`
	PowerLimit                    int       `json:"power_limit"`
	ProtectionLimit               int       `json:"protection_limit"`
	RegenLimit                    int       `json:"regen_limit"`
	StrengthLimit                 int       `json:"strength_limit"`
	IsWeaknessEnabled             bool      `json:"is_weakness_enabled"`
	IsBardPassiveDebuffingEnabled bool      `json:"is_bard_passive_debuffing_enabled"`
	DtrFreezeTimer                int       `json:"dtr_freeze_timer"`
	DtrMax                        float32   `json:"dtr_max"`
	OffPeakLivesNeededAsCents     int       `json:"off_peak_lives_needed_as_cents"`
	Timestamp                     time.Time `json:"timestamp"`
	ServerName                    string    `json:"server_name"`
	GameModeName                  string    `json:"game_mode_name"`
	AttackSpeedName               string    `json:"attack_speed_name"`
	IsInitiallyWhitelisted        bool      `json:"is_initially_whitelisted"`

	CurrentPlayers []onlinePlayer
	Deaths         []death
	Events         []abstractEvent
	Koths          []koth
	SupplyDrops    []supplyDrop
	// Donations      []order TODO
	Messages []ingameMessage
	// Videos         []string TODO
	// TODO -> line items

	Factions []faction
	Bandits  []bandit
}

var currentHcfServerName string
var serverDatas = func() map[string]*serverData {
	serverDatas := make(map[string]*serverData)
	getRowsBlocking("SELECT * FROM get_server_datas()", func(rows pgx.Rows) {
		var serverDataBuffer serverData
		handleFatalPgx(pgx.ForEachRow(rows, []any{&serverDataBuffer.DeathBanMinutes, &serverDataBuffer.WorldBorderRadius, &serverDataBuffer.DefaultKothLootFactor, &serverDataBuffer.SharpnessLimit, &serverDataBuffer.PowerLimit, &serverDataBuffer.ProtectionLimit, &serverDataBuffer.RegenLimit, &serverDataBuffer.StrengthLimit, &serverDataBuffer.IsWeaknessEnabled, &serverDataBuffer.IsBardPassiveDebuffingEnabled, &serverDataBuffer.DtrFreezeTimer, &serverDataBuffer.DtrMax, &serverDataBuffer.OffPeakLivesNeededAsCents, &serverDataBuffer.Timestamp, &serverDataBuffer.ServerName, &serverDataBuffer.GameModeName, &serverDataBuffer.AttackSpeedName, &serverDataBuffer.IsInitiallyWhitelisted}, func() error {
			serverData := serverDataBuffer
			serverDatas[serverDataBuffer.GameModeName+serverDataBuffer.ServerName] = &serverData
			return nil
		}))
	})
	var currentPotentialHcfServerTimestamp time.Time
	for _, data := range serverDatas {
		if strings.Contains(data.GameModeName, "hcf") && data.Timestamp.After(currentPotentialHcfServerTimestamp) && (currentPotentialHcfServerTimestamp.IsZero() || !data.IsInitiallyWhitelisted) {
			currentHcfServerName = data.ServerName
			currentPotentialHcfServerTimestamp = data.Timestamp
		}
	}

	println("serverdatas done")
	return serverDatas
}()

func init() {
	for _, serverData := range serverDatas {
		getRowsBlocking("SELECT * FROM get_7_factions($1, $2)", func(rows pgx.Rows) {
			var faction faction
			handleFatalPgx(pgx.ForEachRow(rows, []any{&faction.Name, &faction.PartyUuid, &faction.FrozenUntil, &faction.CurrentMaxDtr, &faction.CurrentRegenAdjustedDtr}, func() error {
				serverData.Factions = append(serverData.Factions, faction)
				return nil
			}))
		}, serverData.GameModeName, serverData.ServerName)
		getRowsBlocking("SELECT * FROM get_7_newest_bandits($1, $2)", func(rows pgx.Rows) {
			var bandit bandit
			handleFatalPgx(pgx.ForEachRow(rows, []any{&bandit.UserUuid, &bandit.DeathId, &bandit.DeathTimestamp, &bandit.ExpirationTimestamp, &bandit.BanditMessage}, func() error {
				serverData.Bandits = append(serverData.Bandits, bandit)
				return nil
			}))
		}, serverData.GameModeName, serverData.ServerName)
	}

	println("serverdatas init done")
}

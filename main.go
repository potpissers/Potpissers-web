package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
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

	handleRowsBlocking := func(query string, bar func(rows pgx.Rows)) {
		rows, err := postgresPool.Query(context.Background(), query)
		defer rows.Close()
		if err != nil {
			log.Fatal(err)
		}
		bar(rows)
	}

	fetchTips := func(tipsName string) []string {
		var tips []string
		handleRowsBlocking("SELECT tip_message FROM server_tips WHERE server_id = (SELECT id FROM servers WHERE name = '" + tipsName + "')", func(rows pgx.Rows) {
			var tipMessage string
			_, err := pgx.ForEachRow(rows, []any{&tipMessage}, func() error {
				tips = append(tips, tipMessage)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		})
		return tips // TODO -> pass reference and then deference? fuck idk
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
		handleRowsBlocking(Return100NewPlayers, func(rows pgx.Rows) {
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
		handleRowsBlocking(Return16Deaths, func(rows pgx.Rows) {
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
		handleRowsBlocking(Return100Events, func(rows pgx.Rows) {
			var event Event
			_, err := pgx.ForEachRow(rows, []any{&event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
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

//	type Transaction struct {
//	}
//	var transactions []Transaction
//	{
//		handleRowsBlocking(Return100Events, func(rows pgx.Rows) {
//			var event Transaction
//			err := rows.Scan()
//			if err != nil {
//				log.Fatal(err)
//			}
//			transactions = append(transactions, event)
//		})
//	}

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
	}
	serverDatas := make(map[string]*ServerData)
	handleRowsBlocking(ReturnAllServerData, func(rows pgx.Rows) {
		var serverData ServerData
		_, err := pgx.ForEachRow(rows, []any{&serverData.DeathBanMinutes, &serverData.WorldBorderRadius, &serverData.SharpnessLimit, &serverData.PowerLimit, &serverData.ProtectionLimit, &serverData.RegenLimit, &serverData.StrengthLimit, &serverData.IsWeaknessEnabled, &serverData.IsBardPassiveDebuffingEnabled, &serverData.DtrFreezeTimer, &serverData.DtrMax, &serverData.DtrMaxTime, &serverData.DtrOffPeakFreezeTime, &serverData.OffPeakLivesNeededAsCents, &serverData.BardRadius, &serverData.RogueRadius, &serverData.ServerName, &serverData.AttackSpeedName}, func() error {
			serverDatas[serverData.ServerName] = &serverData
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	handleRowsBlocking(ReturnAllOnlinePlayers, func(rows pgx.Rows) {
		var playerName string
		var serverName string
		_, err := pgx.ForEachRow(rows, []any{&playerName, &serverName}, func() error {
			serverDatas[serverName].CurrentPlayers = append(serverDatas[serverName].CurrentPlayers, playerName)
			return nil
		})
		// TODO sort
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/api/servers/", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

//	var chatCache []string // TODO
	http.HandleFunc("/server-chat", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	// TODO handle factions (for each map etc)

	getMainTemplate := func(fileName string) *template.Template {
		mainTemplate, err := template.ParseFiles("main.html", fileName)
		if err != nil {
			log.Fatal(err)
		}
		return mainTemplate
	}
	home, mz, hcf := getMainTemplate("main-home.html"), getMainTemplate("main-mz.html"), getMainTemplate("main-hcf.html")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := home.Execute(w, struct {
			NewPlayers []NewPlayer
			PotpissersTips []string
			Deaths []Death
		}{
			NewPlayers: newPlayers,
			PotpissersTips: potpissersTips,
			Deaths: deaths,
			})
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		err := mz.Execute(w, struct {
			NewPlayers []NewPlayer
			PotpissersTips []string
			Deaths []Death

			AttackSpeed string

			MzTips []string
		}{
			NewPlayers: newPlayers,
			PotpissersTips: potpissersTips,
			Deaths: deaths,

			AttackSpeed: serverDatas["mz"].AttackSpeedName,

			MzTips: mzTips,
			})
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		serverData := serverDatas["hcf"]
		err := hcf.Execute(w, struct {
			NewPlayers []NewPlayer
			PotpissersTips []string
			Deaths []Death

			AttackSpeed string

			DeathBanMinutes int
			OffPeakLivesNeeded float32
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
		}{
			NewPlayers: newPlayers,
			PotpissersTips: potpissersTips,
			Deaths: deaths,

			AttackSpeed: serverData.AttackSpeedName,

			DeathBanMinutes: serverData.DeathBanMinutes,
			OffPeakLivesNeeded: float32(serverData.OffPeakLivesNeededAsCents) / 100.0,
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
			})
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

const Return16Deaths = `SELECT name,
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
LIMIT 16`
const Return100NewPlayers = `SELECT user_uuid, referrer, timestamp, ROW_NUMBER() OVER (ORDER BY timestamp) AS row_number
FROM user_referrals
ORDER BY timestamp
LIMIT 100`
const Return100Events = `SELECT start_timestamp,
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
LIMIT 100`
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
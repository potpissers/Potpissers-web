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
	var postgresPool pgxpool.Pool
	{
		postgresPool, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
		defer postgresPool.Close()
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
		PlayerUuid string
		Referrer  string
		Timestamp time.Time
		RowNumber int
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
		ServerName string
		VictimUserFightId int
		Timestamp time.Time
		VictimUuid string
		// TODO victim inventory
		DeathWorldName string
		DeathX int
		DeathY int
		DeathZ int
		DeathMessage string
		KillerUuid string
		// TODO killer weapon
		// TODO killer inventory
	}
	var deaths []Death
	{
		handleRowsBlocking(Return100Deaths, func(rows pgx.Rows) {
			var death Death
			foo
			err := rows.Scan(&death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil)
			if err != nil {
				log.Fatal(err)
			}
			deaths = append(deaths, death)
		})
	}
	var deathsMu sync.RWMutex
	http.HandleFunc("/api/deaths", func(w http.ResponseWriter, r *http.Request) {
		handlePutJson[Death](r, func(newDeath *Death, r *http.Request) error {return json.NewDecoder(r.Body).Decode(&newDeath)}, &deathsMu, &deaths)
	})

	type Event struct {
		StartTimestamp time.Time
		LootFactor    int
		IsMovementRestricted bool
		CappingUserUUID string
		EndTimestamp time.Time
		CappingPartyUUID string
		World string
		X int
		Y int
		Z int
		ServerName string
		ArenaName string
		Creator string
	}
	var events []Event
	{
		handleRowsBlocking(Return100Events, func(rows pgx.Rows) {
			var event Event
			err := rows.Scan(&event.StartTimestamp, &event.LootFactor, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.World, &event.X, &event.Y, &event.Z, &event.ServerName, &event.ArenaName, &event.Creator)
			if err != nil {
				log.Fatal(err)
			}
			events = append(events, event)
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

	var serverData map[string]map[string]interface{}
	handleRowsBlocking(ReturnAllServerData, func(rows pgx.Rows) {
		var deathBanMinutes int
		var worldBorderRadius int
		var sharpnessLimit int
		var powerLimit int
		var protectionLimit int
		var regenLimit int
		var strengthLimit int
		var isWeaknessEnabled bool
		var isBardPassiveDebuffingEnabled bool
		var dtrFreezeTimer int
		var dtrMax float32
		var dtrMaxTime int
		var dtrOffPeakFreezeTime int
		var offPeakLivesNeededAsCents int
		var bardRadius int
		var rogueRadius int
		var serverName string
		var attackSpeedName string

		err := rows.Scan(&deathBanMinutes, &worldBorderRadius, &sharpnessLimit, &powerLimit, &protectionLimit, &regenLimit, &strengthLimit, &isWeaknessEnabled, &isBardPassiveDebuffingEnabled, &dtrFreezeTimer, &dtrMax, &dtrMaxTime, &dtrOffPeakFreezeTime, &offPeakLivesNeededAsCents, &bardRadius, &rogueRadius, &serverName, &attackSpeedName)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/server-chat", func(w http.ResponseWriter, r *http.Request) {
	})

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
			Tips []string
		}{
			Tips: potpissersTips,
			})
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		err := mz.Execute(w, struct {
			Tips []string
		}{
			Tips: mzTips,
			})
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		err := hcf.Execute(w, struct {
			Tips []string
			ClassInfo []string
		}{
			Tips: cubecoreTips,
			ClassInfo: cubecoreClassTips,
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

const Return100Deaths = `SELECT name,
       victim_user_fight_id,
       Timestamp,
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
ORDER BY Timestamp
LIMIT 100`
const Return100NewPlayers = `SELECT user_uuid, Referrer, Timestamp, ROW_NUMBER() OVER (ORDER BY Timestamp) AS row_number
FROM user_referrals
ORDER BY Timestamp
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

func handlePutJson[T any](r *http.Request, decodeJson func(*T, *http.Request) error, mutex *sync.RWMutex, collection *[]T) {
	if r.Method == "POST" {
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			var newT T
			err := decodeJson(&newT, r)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Lock()
			*collection = append(*collection, newT)
			mutex.Unlock()
		}
	}
}
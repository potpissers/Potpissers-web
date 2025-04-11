package main

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type death struct {
	GameModeName      string      `json:"game_mode_name"`
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

var deaths = getDeathsQuerySlice("SELECT * FROM get_12_latest_network_deaths()")

func init() {
	for _, serverData := range serverDatas {
		serverData.Deaths = getDeathsQuerySlice("SELECT * FROM get_12_latest_server_deaths($1, $2)", serverData.GameModeName, serverData.ServerName)
	}
	println("deaths done")
}

func getDeathsQuerySlice(query string, params ...any) []death {
	var deaths []death
	getRowsBlocking(query, func(rows pgx.Rows) {
		var death death
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.GameModeName, &death.ServerName, &death.VictimUserFightId, &death.Timestamp, &death.VictimUuid, nil, &death.DeathWorldName, &death.DeathX, &death.DeathY, &death.DeathZ, &death.DeathMessage, &death.KillerUuid, nil, nil}, func() error {
			deaths = append(deaths, death)
			return nil
		}))
	}, params...)
	return deaths
}

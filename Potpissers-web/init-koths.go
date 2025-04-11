package main

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type koth struct {
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
	GameModeName         string    `json:"game_mode_name"`
	ServerName           string    `json:"server_name"`
	ArenaName            string    `json:"arena_name"`
	Creator              string    `json:"creator"`
}

var koths = getArenaQuerySlice("SELECT * FROM get_14_newest_network_koths()")

func init() {
	for _, serverData := range serverDatas {
		serverData.Koths = getArenaQuerySlice("SELECT * FROM get_14_newest_server_koths($1, $2)", serverData.GameModeName, serverData.ServerName)
	}
	println("koths done")
}

func getArenaQuerySlice(query string, params ...any) []koth {
	var events []koth
	getRowsBlocking(query, func(rows pgx.Rows) {
		var koth koth
		handleFatalPgx(pgx.ForEachRow(rows, []any{&koth.ServerKothsId, &koth.StartTimestamp, &koth.LootFactor, &koth.MaxTimer, &koth.IsMovementRestricted, &koth.CappingUserUUID, &koth.EndTimestamp, &koth.CappingPartyUUID, &koth.CapMessage, &koth.World, &koth.X, &koth.Y, &koth.Z, &koth.GameModeName, &koth.ServerName, &koth.ArenaName, &koth.Creator}, func() error {
			events = append(events, koth)
			return nil
		}))
	}, params...)
	return events
}

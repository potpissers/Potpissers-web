package main

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type event struct {
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

var events = func() []event {
	var events []event
	getRowsBlocking("SELECT * FROM get_14_newest_network_koths()", func(rows pgx.Rows) {
		var event event
		handleFatalPgx(pgx.ForEachRow(rows, []any{&event.ServerKothsId, &event.StartTimestamp, &event.LootFactor, &event.MaxTimer, &event.IsMovementRestricted, &event.CappingUserUUID, &event.EndTimestamp, &event.CappingPartyUUID, &event.CapMessage, &event.World, &event.X, &event.Y, &event.Z, &event.GameModeName, &event.ServerName, &event.ArenaName, &event.Creator}, func() error {
			events = append(events, event)
			return nil
		}))
	})
	println("events done")
	return events
}()
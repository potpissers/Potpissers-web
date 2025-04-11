package main

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type supplyDrop struct {
	Id             int       `json:"id"`
	ServerName     string    `json:"server_name"`
	GameModeName   string    `json:"game_mode_name"`
	StartTimestamp time.Time `json:"start_timestamp"`

	NullableServerKothId     pgtype.Int4 `json:"nullable_server_koth_id"`
	IsKothMovementRestricted bool        `json:"is_koth_movement_restricted"`
	WorldName                string      `json:"world_name"`
	X                        int         `json:"x"`
	Y                        int         `json:"y"`
	Z                        int         `json:"z"`
	Radius                   int         `json:"radius"`
	ChestOpenTimestamp       time.Time   `json:"chest_open_timestamp"`
	LootFactor               int         `json:"loot_factor"`
	RestockTimer             int         `json:"restock_timer"`
	RestockAmount            int         `json:"restock_amount"`

	EndTimestamp time.Time `json:"end_timestamp"`
	WinMessage   string    `json:"win_message"`
}

var supplyDrops = getSupplyDropsQuerySlice("SELECT * FROM get_14_newest_network_koths()")

func init() {
	for _, serverData := range serverDatas {
		serverData.SupplyDrops = getSupplyDropsQuerySlice("SELECT * FROM get_14_newest_network_supply_drops($1, $2)", serverData.ServerName, serverData.GameModeName)
	}
	print("supply drops done")
}

func getSupplyDropsQuerySlice(query string, params ...any) []supplyDrop {
	var events []supplyDrop
	getRowsBlocking(query, func(rows pgx.Rows) {
		var drop supplyDrop
		handleFatalPgx(pgx.ForEachRow(rows, []any{&drop.Id, &drop.ServerName, &drop.GameModeName, &drop.StartTimestamp, &drop.NullableServerKothId, &drop.IsKothMovementRestricted, &drop.WorldName, &drop.X, &drop.Y, &drop.Z, &drop.Radius, &drop.ChestOpenTimestamp, &drop.LootFactor, &drop.RestockTimer, &drop.RestockAmount, &drop.EndTimestamp, &drop.WinMessage}, func() error {
			events = append(events, drop)
			return nil
		}))
	}, params)
	return events
}

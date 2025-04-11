package main

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type onlinePlayer struct {
	Uuid          string    `json:"uuid"`
	Name          string    `json:"name"`
	GameModeName  string    `json:"game_mode_name"`
	ServerName    string    `json:"server_name"`
	ActiveFaction *string   `json:"active_faction"`
	NetworkJoin   time.Time `json:"network_join"`
	ServerJoin    time.Time `json:"server_join"`
}

var networkPlayers = func() []onlinePlayer {
	var currentPlayers = []onlinePlayer{}
	getRowsBlocking("SELECT * FROM get_online_players()", func(rows pgx.Rows) {
		var t onlinePlayer
		handleFatalPgx(pgx.ForEachRow(rows, []any{&t.Uuid, &t.Name, &t.GameModeName, &t.ServerName, &t.ActiveFaction, &t.NetworkJoin, &t.ServerJoin}, func() error {
			currentPlayers = append(currentPlayers, t) // TODO sort names
			serverData := serverDatas[t.GameModeName+t.ServerName]
			serverData.CurrentPlayers = append(serverData.CurrentPlayers, t)
			return nil
		}))
	})

	println("online players done")
	return currentPlayers
}()

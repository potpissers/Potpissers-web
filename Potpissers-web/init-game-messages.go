package main

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type ingameMessage struct {
	Uuid         string    `json:"uuid"`
	Message      string    `json:"message"`
	GameModeName string    `json:"game_mode_name"`
	ServerName   string    `json:"server_name"`
	Timestamp    time.Time `json:"timestamp"`
}

var messages = getIngameMessagesQuerySlice("SELECT * FROM get_14_newest_network_messages()")

func init() {
	for _, serverData := range serverDatas {
		serverData.Messages = getIngameMessagesQuerySlice("SELECT * FROM get_14_newest_server_messages($1, $2)", serverData.GameModeName, serverData.ServerName)
	}
	print("ingame messages done")
}

func getIngameMessagesQuerySlice(query string, params ...any) []ingameMessage {
	var events []ingameMessage
	getRowsBlocking(query, func(rows pgx.Rows) {
		var drop ingameMessage
		handleFatalPgx(pgx.ForEachRow(rows, []any{&drop.Uuid, &drop.Message, &drop.GameModeName, &drop.ServerName, &drop.Timestamp}, func() error {
			events = append(events, drop)
			return nil
		}))
	}, params)
	return events
}

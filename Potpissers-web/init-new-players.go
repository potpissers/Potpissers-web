package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type newPlayer struct {
	PlayerUuid string    `json:"player_uuid"`
	Referrer   *string   `json:"referrer"`
	Timestamp  time.Time `json:"timestamp"`
	RowNumber  int       `json:"row_number"`

	PlayerName string `json:"player_name"`
}

var newPlayers = func() []newPlayer {
	var newPlayers []newPlayer
	getRowsBlocking("SELECT * FROM get_10_newest_players()", func(rows pgx.Rows) {
		var death newPlayer
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.PlayerUuid, &death.Referrer, &death.Timestamp, &death.RowNumber}, func() error {
			newPlayers = append(newPlayers, death)
			return nil
		}))
	})

	for i := range newPlayers {
		resp, err := http.Get("https://api.minecraftservices.com/minecraft/profile/lookup/" + newPlayers[i].PlayerUuid)
		if err != nil {
			log.Fatal(err)
		}
		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		newPlayers[i].PlayerName = result["name"].(string)
	}

	println("new players done")
	return newPlayers
}()
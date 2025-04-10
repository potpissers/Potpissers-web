package main

type ingameMessage struct {
	Uuid         string `json:"uuid"`
	Message      string `json:"message"`
	GameModeName string `json:"game_mode_name`
	ServerName   string `json:"server_name"`
}

var messages []ingameMessage

// TODO -> init grab postgres

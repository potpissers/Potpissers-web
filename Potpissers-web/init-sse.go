package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

func init() {
	connection, err := postgresPool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// defer'd -> below goroutine
	for _, channelName := range []string{"referrals", "deaths", "events", "online", "chat", "server_data"} {
		_, err = connection.Exec(context.Background(), "LISTEN "+channelName)
		if err != nil {
			connection.Release()
			log.Fatal(err)
		}
	}
	go func() {
		defer connection.Release()
		for {
			notification, err := connection.Conn().WaitForNotification(context.Background())
			if err != nil {
				log.Fatal(err)
			}
			switch notification.Channel { // TODO -> transitioning from ssr to csr like this absolutely can desync, oh well
			case "referrals":
				{ // TODO -> NVM just block the response if csr is happening
					var t newPlayer
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"referrals", t})
					if err != nil {
						log.Fatal(err)
					}

					newPlayers = append([]newPlayer{t}, newPlayers...)

					home = getMainTemplateBytes(homeTemplate, "hub")
					mz = getMainTemplateBytes(mzTemplate, "mz")
					hcf = getMainTemplateBytes(hcfTemplate, "hcf")

					handleSseData(jsonBytes, homeConnections, mzConnections, hcfConnections)
				}
			case "deaths":
				{
					var t death
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"deaths", t})
					handleFatalErr(err)
					serverData := serverDatas[t.GameModeName+t.ServerName]
					handleServerDataJsonPrepend[death](&deaths, t, jsonBytes, &serverData.deaths, serverData.gameModeName)
				}
			case "events":
				{
					var t event
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"events", t})
					handleFatalErr(err)
					serverData := serverDatas[t.GameModeName+t.ServerName]
					handleServerDataJsonPrepend[event](&events, t, jsonBytes, &serverData.events, serverData.gameModeName)
				}
			case "chat":
				{
					var t ingameMessage
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"chat", t})
					handleFatalErr(err)
					serverData := serverDatas[t.GameModeName+t.ServerName]
					handleServerDataJsonPrepend[ingameMessage](&messages, t, jsonBytes, &serverData.messages, serverData.gameModeName)
				}
			case "online":
				{
					var t onlinePlayer
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"online", t})
					handleFatalErr(err)
					serverData := serverDatas[t.GameModeName+t.ServerName]
					handleServerDataJsonPrepend[onlinePlayer](&currentPlayers, t, jsonBytes, &serverData.currentPlayers, serverData.gameModeName)
				}
			case "offline":
				{
					var t onlinePlayer
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"offline", t})
					handleFatalErr(err)

					for i, data := range currentPlayers {
						if data == t {
							currentPlayers = append(currentPlayers[:i], currentPlayers[i+1:]...)
							break
						}
					}
					home = getMainTemplateBytes(homeTemplate, "hub")
					handleSseData(jsonBytes, homeConnections)

					serverData := serverDatas[t.GameModeName+t.ServerName]
					for i, data := range serverData.currentPlayers {
						if data == t {
							serverData.currentPlayers = append(serverData.currentPlayers[:i], serverData.currentPlayers[i+1:]...)
							break
						}
					}
					switch serverData.gameModeName {
					case "hcf":
						{
							hcf = getMainTemplateBytes(hcfTemplate, "hcf")
							handleSseData(jsonBytes, hcfConnections)
						}
					case "mz":
						{
							mz = getMainTemplateBytes(mzTemplate, "mz")
							handleSseData(jsonBytes, mzConnections)
						}
					}
				}
			case "bandits":
				{
				} //TODO
			case "server_data":
				{
				} // TODO
			default:
				log.Fatal("postgres listen err")
			}
		}
	}()
}

type sseConnectionsData struct {
	mop   map[*struct{}]chan []byte
	mutex *sync.RWMutex
}

var homeConnections = sseConnectionsData{make(map[*struct{}]chan []byte), &sync.RWMutex{}}
var mzConnections = sseConnectionsData{make(map[*struct{}]chan []byte), &sync.RWMutex{}}
var hcfConnections = sseConnectionsData{make(map[*struct{}]chan []byte), &sync.RWMutex{}}

type sseMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func handleSseData(bytes []byte, sseConnectionMaps ...sseConnectionsData) {
	for _, mop := range sseConnectionMaps {
		go func(data sseConnectionsData) {
			data.mutex.RLock()
			for _, ch := range data.mop {
				ch <- []byte("data: " + string(bytes) + "\n\n")
			}
			data.mutex.RUnlock()
		}(mop)
	}
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func init() {
	connection, err := postgresPool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// defer'd -> below goroutine
	for _, channelName := range []string{"deaths", "drops", "chat", "koths", "referrals", "online", "offline", "server_data", "bandits", "factions"} {
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
			case "deaths": // TODO -> NVM just block the response if csr is happening
				{
					var t death
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"deaths", t})
					handleFatalErr(err)
					handleServerDataJsonPrepend[death](&deaths, t, jsonBytes, &serverDatas[t.GameModeName+t.ServerName].Deaths)
				}
			case "drops":
				{ // TODO
				}
			case "chat":
				{
					var t ingameMessage
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"chat", t})
					handleFatalErr(err)
					handleServerDataJsonPrepend[ingameMessage](&messages, t, jsonBytes, &serverDatas[t.GameModeName+t.ServerName].Messages)
				}
			case "koths":
				{
					var t koth
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"koths", t})
					handleFatalErr(err)
					handleServerDataJsonPrepend[koth](&koths, t, jsonBytes, &serverDatas[t.GameModeName+t.ServerName].Koths)
				}
			case "referrals":
				{
					var t newPlayer
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"referrals", t})
					handleFatalErr(err)

					newPlayers = append([]newPlayer{t}, newPlayers...)

					home = getMainTemplateBytes("hub")
					mz = getMainTemplateBytes("mz")
					hcf = getMainTemplateBytes("hcf")

					handleSseData(jsonBytes, mainConnections)
				}
			case "online":
				{
					var t onlinePlayer
					handleFatalErr(json.Unmarshal([]byte(notification.Payload), &t))
					jsonBytes, err := json.Marshal(sseMessage{"online", t})
					handleFatalErr(err)
					handleServerDataJsonPrepend[onlinePlayer](&currentPlayers, t, jsonBytes, &serverDatas[t.GameModeName+t.ServerName].CurrentPlayers)
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
					serverData := serverDatas[t.GameModeName+t.ServerName]
					for i, data := range serverData.CurrentPlayers {
						if data == t {
							serverData.CurrentPlayers = append(serverData.CurrentPlayers[:i], serverData.CurrentPlayers[i+1:]...)
							break
						}
					}

					home = getMainTemplateBytes("hub")
					hcf = getMainTemplateBytes("hcf")
					mz = getMainTemplateBytes("mz")
					handleSseData(jsonBytes, mainConnections)
				}
			case "server_data":
				{
				} // TODO
			case "bandits":
				{
				} //TODO
			case "factions":
				{ // TODO
				}
			default:
				log.Fatal("postgres listen err")
			}
		}
	}()

	for _, endpoint := range []string{"/", "hub", "mz", "hcf"} {
		http.HandleFunc("/api/sse"+endpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			flusher, ok := w.(http.Flusher)
			if ok {
				flusher.Flush()

				ch := make(chan []byte, 2)
				pointer := &struct{}{}
				mutex := mainConnections.mutex

				mutex.Lock()
				mainConnections.mop[pointer] = ch
				mutex.Unlock()

			whileTrue:
				for {
					select {
					case msg := <-ch:
						{
							_, err := w.Write(msg)
							if err != nil {
								break whileTrue
							}
							flusher.Flush()
						}
					case <-r.Context().Done():
						{
							break whileTrue
						}
					}
				}

				mutex.Lock()
				mainConnections.mop[pointer] = nil
				mutex.Unlock()
			}
		})
	}
}

type sseConnectionsData struct {
	mop   map[*struct{}]chan []byte
	mutex *sync.RWMutex
}

var mainConnections = sseConnectionsData{make(map[*struct{}]chan []byte), &sync.RWMutex{}}

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
func handleServerDataJsonPrepend[T any](homeSlice *[]T, t T, bytes []byte, serverSlice *[]T) {
	*homeSlice = append([]T{t}, *homeSlice...) // TODO -> this is necessary because html/css and go's templating can't handle reversing it for some reason. go's templater could maybe do it but it seems like more processing than this takes
	home = getMainTemplateBytes("hub")
	hcf = getMainTemplateBytes("hcf")
	mz = getMainTemplateBytes("mz")
	handleSseData(bytes, mainConnections)

	*serverSlice = append([]T{t}, *serverSlice...)
}

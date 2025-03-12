package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"net/http"
	"strings"
	"sync"
)
func handleFatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func handleFatalPgx(_ pgconn.CommandTag, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func getFatalJsonT[T any](request *http.Request) T {
	resp, err := (&http.Client{}).Do(request)
	handleFatalErr(err)
	defer resp.Body.Close()
	var messages T
	handleFatalErr(json.NewDecoder(resp.Body).Decode(&messages))
	return messages
}

func getRowsBlocking(query string, bar func(rows pgx.Rows), params ...any) {
	rows, err := postgresPool.Query(context.Background(), query, params...)
	defer rows.Close()
	handleFatalErr(err)
	bar(rows)
}

func handleLocalhostJsonPatch[T any](r *http.Request, decodeJson func(*T, *http.Request) error, mutex *sync.RWMutex, collection *[]T) {
	if r.Method == "PATCH" {
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			var newT T
			handleFatalErr(decodeJson(&newT, r))

			mutex.Lock()
			*collection = append([]T{newT}, *collection...) // TODO -> this is necessary because html/css and go's templating can't handle reversing it for some reason. go's templater could maybe do it but it seems like more processing than this takes
			mutex.Unlock()
		}
	}
}
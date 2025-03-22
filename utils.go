package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"io"
	"log"
	"net/http"
	"os"
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
func handleGetFatalJsonT[T any](request *http.Request) T {
	resp, err := (&http.Client{}).Do(request)
	handleFatalErr(err)
	defer resp.Body.Close()
	return getFatalJsonT[T](resp)
}

func getFatalJsonT[T any](resp *http.Response) T {
	var messages T
	handleFatalErr(json.NewDecoder(resp.Body).Decode(&messages))
	return messages
}

func handleSseData(sseConnections []sseConnection, bytes []byte) {
	for _, data := range sseConnections {
		go func() {
			_, err := data.response.Write(bytes)
			if err != nil {
				log.Println(err)
			} else {
				data.flusher.Flush()
			}
		}()
	}
}

func getMojangApiUuidRequest(username string) (*http.Response, error) {
	return http.Get(minecraftUsernameLookupUrl + username)
}

func getRowsBlocking(query string, bar func(rows pgx.Rows), params ...any) {
	rows, err := postgresPool.Query(context.Background(), query, params...)
	defer rows.Close()
	handleFatalErr(err)
	bar(rows)
}

func getFatalRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	handleFatalErr(err)
	return req
}
func addSquareHeaders(request *http.Request) {
	request.Header.Add("Authorization", "Bearer "+ os.Getenv("SQUARE_ACCESS_TOKEN"))
	request.Header.Add("Content-Type", "application/json")
}
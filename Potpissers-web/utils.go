package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return getFatalJsonT[T](resp)
}

func getFatalJsonT[T any](resp *http.Response) T {
	var messages T
	//	println(resp.StatusCode)
	handleFatalErr(json.NewDecoder(resp.Body).Decode(&messages))
	return messages
}

func getMojangApiUuidRequest(username string) (*http.Response, error) {
	return http.Get(minecraftUsernameLookupUrl + username)
}

func getRowsBlocking(query string, bar func(rows pgx.Rows), params ...any) {
	rows, err := postgresPool.Query(context.Background(), query, params...)
	handleFatalErr(err)
	defer rows.Close()
	bar(rows)
}

func getFatalRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	handleFatalErr(err)
	return req
}

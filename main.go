package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	connect, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
	if err != nil {
		log.Fatal("hey")
	}
	err = connect.Ping(context.Background())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("main.html", "main-home.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		template.ExecuteTemplate(w, "main.html", nil)
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("main.html", "main-hcf.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		template.ExecuteTemplate(w, "main.html", nil)
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("main.html", "main-mz.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		template.ExecuteTemplate(w, "main.html", nil)
	})

	http.Handle("/style.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	http.ListenAndServe(":443", nil)
}

package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if (err != nil) {
		log.Fatal("INVALID ENVIRONMENT VARIABLE")
	}
	config := pgxpool.Config{}
	config.ConnConfig.Host = os.Getenv("POSTGRES_HOST")
	config.ConnConfig.User = os.Getenv("POSTGRES_USERNAME")
	config.ConnConfig.Password = os.Getenv("POSTGRES_PASSWORD")
	config.ConnConfig.Port = uint16(port)
	connect, err := pgxpool.NewWithConfig(context.Background(), &config)
	if err != nil {
		return
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

	http.ListenAndServe(":8080", nil)
}

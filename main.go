package main

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
)

var postgresConnection *pgxpool.Pool
func init() {
	var err error
	postgresConnection, err = pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
	if err != nil {
		log.Fatal("hey")
	}
}

func main() {
	var preparedHcfContentHtml string
	var preparedMinezContentHtml string

	cubecoreTips := fetchTips("cubecore")
	mzTips := fetchTips("minez")

	home := getMainTemplate("main-home.html")
	hcf := getMainTemplate("main-hcf.html")
	mz := getMainTemplate("main-mz.html")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleMainTemplate(home, w)
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		handleMainTemplate(hcf, w)
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		handleMainTemplate(mz, w)
	})
	http.Handle("/style.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem",  "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil)
}
func getMainTemplate(fileName string, tipsName string) string {
	hey, err := template.ParseFiles("main.html", fileName)
	if err != nil {
		log.Fatal("template error")
	}
	var buffer bytes.Buffer

	err = hey.ExecuteTemplate(&buffer, "main.html", struct {
		Tips []string
	}{
		Tips: fetchTips(tipsName),
	})
	if err != nil {
		log.Fatal("template error1")
	}
	return buffer.String()
}
func fetchTips(tipsName string) []string {
	rows, err := postgresConnection.Query(context.Background(), "SELECT tip_message FROM server_tips WHERE server_id = (SELECT id FROM servers WHERE name = '" + tipsName + "')")
	if err != nil {
		log.Fatal("cubecoreTips1")
	}
	var cubecoreTips []string
	for rows.Next() {
		var tipMessage string
		err = rows.Scan(&tipMessage);
		if err != nil {
			log.Fatal("cubecoreTips2")
		}
		cubecoreTips = append(cubecoreTips, tipMessage)
	}
	rows.Close()
	return cubecoreTips
}

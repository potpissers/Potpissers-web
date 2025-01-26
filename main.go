package main

import (
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
		log.Fatal(err)
	}
}

func main() {
	cubecoreTips := fetchTips("cubecore")
	mzTips := fetchTips("minez")

	home := getMainTemplate("main-home.html")
	hcf := getMainTemplate("main-hcf.html")
	mz := getMainTemplate("main-mz.html")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := home.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		executeMainTipsTemplate(hcf, w, cubecoreTips)
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		executeMainTipsTemplate(mz, w, mzTips)
	})
	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	err := http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem",  "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil)
	if (err != nil) {
		log.Fatal(err)
	}
}
func fetchTips(tipsName string) []string {
	rows, err := postgresConnection.Query(context.Background(), "SELECT tip_message FROM server_tips WHERE server_id = (SELECT id FROM servers WHERE name = '" + tipsName + "')")
	if err != nil {
		log.Fatal(err)
	}
	var cubecoreTips []string
	for rows.Next() {
		var tipMessage string
		err = rows.Scan(&tipMessage);
		if err != nil {
			log.Fatal(err)
		}
		cubecoreTips = append(cubecoreTips, tipMessage)
	}
	rows.Close()
	return cubecoreTips
}
func getMainTemplate(fileName string) *template.Template {
	hey, err := template.ParseFiles("main.html", fileName)
	if err != nil {
		log.Fatal(err)
	}
	return hey
}
func executeMainTipsTemplate(template2 *template.Template, w http.ResponseWriter, tips []string) {
	err := template2.Execute(w, struct {
		Tips []string
	}{
		Tips: tips,
	})
	if (err != nil) {
		log.Fatal(err)
	}
}

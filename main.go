package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
)

var PostgresPoolFinal *pgxpool.Pool
func init() {
	var err error
	PostgresPoolFinal, err = pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	potpissersTips := fetchTips("null")
	cubecoreTips := append(fetchTips("cubecore"), potpissersTips...)
	mzTips := append(fetchTips("minez"), potpissersTips...)
	cubecoreClassTips := fetchTips("cubecore_classes")

//	TODO deaths

	home := getMainTemplate("main-home.html")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := home.Execute(w, struct {
			Tips []string
		}{
			Tips: potpissersTips,
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	mz := getMainTemplate("main-mz.html")
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		err := mz.Execute(w, struct {
			Tips []string
		}{
			Tips: mzTips,
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	hcf := getMainTemplate("main-hcf.html")
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		err := hcf.Execute(w, struct {
			Tips []string
			ClassInfo []string
		}{
			Tips: cubecoreTips,
			ClassInfo: cubecoreClassTips,
		})
		if err != nil {
			log.Fatal(err)
		}
	})

	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	err := http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem",  "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
func fetchTips(tipsName string) []string {
	rows, err := PostgresPoolFinal.Query(context.Background(), "SELECT tip_message FROM server_tips WHERE server_id = (SELECT id FROM servers WHERE name = '" + tipsName + "')")
	if err != nil {
		log.Fatal(err)
	}
	var cubecoreTips []string
	for rows.Next() {
		var tipMessage string
		err = rows.Scan(&tipMessage)
		if err != nil {
			log.Fatal(err)
		}
		cubecoreTips = append(cubecoreTips, tipMessage)
	}
	rows.Close()
	return cubecoreTips
}
func fetchDeaths(tipsName string) []string {
	rows, err := PostgresPoolFinal.Query(context.Background(), "SELECT tip_message FROM server_tips WHERE server_id = (SELECT id FROM servers WHERE name = '" + tipsName + "')")
	if err != nil {
		log.Fatal(err)
	}
	var cubecoreTips []string
	for rows.Next() {
		var tipMessage string
		err = rows.Scan(&tipMessage)
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

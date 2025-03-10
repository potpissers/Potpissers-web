package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)
var postgresPool = func() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}
	return pool
	// defer'd => main
}()

func main() {
	defer postgresPool.Close()

	mojangUsernameProxyEndpoint := "/api/proxy/mojang/username/"
	http.HandleFunc(mojangUsernameProxyEndpoint, func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("https://api.minecraftservices.com/minecraft/profile/lookup/name/" + strings.TrimPrefix(r.URL.Path, mojangUsernameProxyEndpoint))
		if err != nil {
			log.Println(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(resp.Body)

		// TODO -> headers ?
		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Fatal(w.Write(home))
	})
	http.HandleFunc("/mz", func(w http.ResponseWriter, r *http.Request) {
		log.Fatal(w.Write(mz))
	})
	http.HandleFunc("/hcf", func(w http.ResponseWriter, r *http.Request) {
		log.Fatal(w.Write(hcf))
	})
	http.HandleFunc("/github", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/potpissers", http.StatusMovedPermanently)
	})
	http.HandleFunc("/reddit", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.reddit.com/r/potpissers/", http.StatusMovedPermanently)
	})
	http.HandleFunc("/discord", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://discord.gg/Cqnvktf7EF", http.StatusFound)
	})

	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	http.Handle("/static-donate.js", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	log.Fatal(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem",  "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil))
}
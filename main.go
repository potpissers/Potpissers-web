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
	handleFatalErr(err)
	return pool
	// defer'd => main
}()

func main() {
	defer postgresPool.Close()

	const mojangUsernameProxyEndpoint = "/api/proxy/mojang/username/"
	http.HandleFunc(mojangUsernameProxyEndpoint, func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(minecraftUsernameLookupUrl + strings.TrimPrefix(r.URL.Path, mojangUsernameProxyEndpoint))
		if err != nil {
			log.Println(err)
			return
		} else {
			defer resp.Body.Close()

			// TODO -> headers ?
			w.WriteHeader(resp.StatusCode)
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Println(err)
				return
			}
		}
	})

	home = getHome()
	hcf = getHcf()
	mz = getMz()

	for _, data := range []struct {
		endpoint string
		bytes    []byte
	}{
		{endpoint: "/", bytes: home},
		{endpoint: "/hub", bytes: home},
		{endpoint: "/mz", bytes: mz},
//		{endpoint: "/kollusion", bytes: kollusion}, // TODO
		{endpoint: "/hcf", bytes: hcf},
//		{endpoint: "/cubecore", bytes: cubecore},
	} {
		http.HandleFunc(data.endpoint, func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(data.bytes)
			handleFatalErr(err)
		})
	}

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

	log.Println(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem", "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil))
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const minecraftUsernameLookupUrl = "https://api.minecraftservices.com/minecraft/profile/lookup/name/"
const potpissersRedditApiUrl = "https://oauth.reddit.com/r/potpissers/new.json?limit=100"
const frontendDirName = "./frontend"

var postgresPool = func() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING"))
	handleFatalErr(err)
	// defer'd => main
	return pool
}()

func main() {
	defer postgresPool.Close()
	println("init done")

	doApi()
	println("main api done")

	home = getMainTemplateBytes(homeTemplate, "hub")
	println("home template done")
	hcf = getMainTemplateBytes(hcfTemplate, "hcf")
	println("hcf template done")
	mz = getMainTemplateBytes(mzTemplate, "mz")
	println("mz template done")

	http.HandleFunc("/github", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/potpissers", http.StatusMovedPermanently)
	})
	http.HandleFunc("/reddit", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.reddit.com/r/potpissers/", http.StatusMovedPermanently)
	})
	http.HandleFunc("/discord", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://discord.gg/Cqnvktf7EF", http.StatusFound)
	})

	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/static-donate.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/static-init.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/favicon.png", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/mz-map/", http.StripPrefix("/mz-map", http.FileServer(http.Dir(frontendDirName+"/mz-map"))))

	println("starting server")
	log.Println(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem", "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil))
}

func handleServerDataJsonPrepend[T any](homeSlice *[]T, t T, bytes []byte, serverSlice *[]T, gameModeName string) {
	*homeSlice = append([]T{t}, *homeSlice...) // TODO -> this is necessary because html/css and go's templating can't handle reversing it for some reason. go's templater could maybe do it but it seems like more processing than this takes
	home = getMainTemplateBytes(homeTemplate, "hub")
	handleSseData(bytes, homeConnections)

	*serverSlice = append([]T{t}, *serverSlice...)
	switch gameModeName {
	case "hcf":
		{
			hcf = getMainTemplateBytes(hcfTemplate, "hcf")
			handleSseData(bytes, hcfConnections)
		}
	case "mz":
		{
			mz = getMainTemplateBytes(mzTemplate, "mz")
			handleSseData(bytes, mzConnections)
		}
	}
}

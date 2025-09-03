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
const frontendDirName = "./frontend/"

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

	home = getMainTemplateBytes("hub")
	println("home template done")
	hcf = getMainTemplateBytes("hcf" + currentHcfServerName)
	println("hcf template done")
	mz = getMainTemplateBytes("mz")
	println("mz template done")

	for endpoint, bytes := range map[string]*[]byte{
		"/":    &home,
		"/hub": &home,
		"/mz":  &mz,
		"/hcf": &hcf,
	} {
		pointer := bytes
		http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", "frame-ancestors *")
			_, err := w.Write(*pointer)
			handleFatalErr(err)

			handleRedditPostDataUpdate()
			handleDiscordMessagesUpdate(discordGeneralChan, discordGeneralChannelId, &mostRecentDiscordGeneralMessageId, &discordMessages, "general")
			handleDiscordMessagesUpdate(discordChangelogChan, discordChangelogChannelId, &mostRecentDiscordChangelogMessageId, &changelog, "changelog")
			handleDiscordMessagesUpdate(discordAnnouncementsChan, discordAnnouncementsChannelId, &mostRecentDiscordAnnouncementsMessageId, &announcements, "announcements")
		})
		println(endpoint + " done")
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

	http.Handle("/static.css", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/static.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/static-donate.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/static-init.js", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/potpisser.jpg", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))
	http.Handle("/favicon.png", http.StripPrefix("/", http.FileServer(http.Dir(frontendDirName))))

	http.Handle("/mz-map/", http.StripPrefix("/mz-map", http.FileServer(http.Dir(frontendDirName+"/mz-map"))))

	println("starting server")
	for { // TODO fix whatever the fuck is causing EOF error
		err := http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/potpissers.com/fullchain.pem", "/etc/letsencrypt/live/potpissers.com/privkey.pem", nil)
		if err != nil {
			log.Println(err)
		}
	}
}

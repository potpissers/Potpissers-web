package main

import (
	"bytes"
	"html/template"
	"math/rand"
	"time"
)

type serverTemplateData struct {
	CurrentPlayers []onlinePlayer
}

var mainTemplate = func() *template.Template {
	mainTemplate, err := template.ParseFiles(
		frontendDirName+"/index.gohtml",
		frontendDirName+"index-videos.gohtml",
		frontendDirName+"index-main.gohtml",
		frontendDirName+"index-main-top.gohtml",
		frontendDirName+"index-main-bottom.gohtml",
		frontendDirName+"index-events.gohtml",
		frontendDirName+"index-donations.gohtml",
		frontendDirName+"index-deaths.gohtml",
		frontendDirName+"index-content.gohtml",
		frontendDirName+"index-content-side-title.gohtml",
		frontendDirName+"index-content-side-body.gohtml",
		frontendDirName+"index-content-main-title.gohtml",
		frontendDirName+"index-content-main-body.gohtml")
	handleFatalErr(err)
	println("main template done")
	return mainTemplate
}()

func getMainTemplateBytes(gameModeName string) []byte {
	serverTemplateDatas := make(map[string]serverTemplateData)
	for gameModeName := range serverDatas {
		var serverTemplateData serverTemplateData
		for _, currentPlayer := range networkPlayers {
			if gameModeName == currentPlayer.GameModeName {
				serverTemplateData.CurrentPlayers = append(serverTemplateData.CurrentPlayers, currentPlayer)
			}
		}
		serverTemplateDatas[gameModeName] = serverTemplateData
	}

	var buffer bytes.Buffer
	donationsMu.RLock()
	handleFatalErr(mainTemplate.Execute(&buffer, struct {
		GameModeName       string
		BackgroundImageUrl redditImagePost
		NetworkPlayers     []onlinePlayer
		ServerDatas        map[string]*serverData
		NewPlayers         []newPlayer
		ContentData        map[gameModeTips][]tip

		Messages             []ingameMessage
		Announcements        []discordMessage
		Changelog            []discordMessage
		DiscordMessages      []discordMessage
		Donations            []order
		LineItemData         []lineItemData
		RedditVideos         []redditVideoPost
		DiscordId            string
		CurrentHcfServerName string
		Deaths               []death
		Events               []abstractEvent
		Koths                []koth
		SupplyDrops          []supplyDrop

		ServerTemplateDatas map[string]serverTemplateData
	}{
		GameModeName:       gameModeName,
		BackgroundImageUrl: redditImagePosts[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(redditImagePosts))],
		NetworkPlayers:     networkPlayers,
		ServerDatas:        serverDatas,
		NewPlayers:         newPlayers,
		ContentData:        contentData,

		Messages:             messages,
		Announcements:        announcements,
		Changelog:            changelog,
		DiscordMessages:      discordMessages,
		Donations:            donations,
		LineItemData:         lineItemDatas,
		RedditVideos:         redditVideoPosts,
		DiscordId:            "1245300045188956252",
		CurrentHcfServerName: currentHcfServerName,
		Deaths:               deaths,
		Events:               events,
		Koths:                koths,
		SupplyDrops:          supplyDrops,

		ServerTemplateDatas: serverTemplateDatas,
	}))
	donationsMu.RUnlock()
	return buffer.Bytes()
}

var home []byte
var mz []byte
var hcf []byte

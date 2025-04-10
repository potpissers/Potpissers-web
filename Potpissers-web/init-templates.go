package main

import (
	"bytes"
	"html/template"
	"math/rand"
	"time"
)

var homeTemplate *template.Template
var mzTemplate *template.Template
var hcfTemplate *template.Template

func init() {
	getMainTemplate := func(fileName string) *template.Template {
		mainTemplate, err := template.ParseFiles(frontendDirName+"/main.gohtml", fileName)
		handleFatalErr(err)
		return mainTemplate
	}

	homeTemplate = getMainTemplate(frontendDirName + "/main-home.gohtml")
	mzTemplate = getMainTemplate(frontendDirName + "/main-mz.gohtml")
	hcfTemplate = getMainTemplate(frontendDirName + "/main-hcf.gohtml")
}

func getMainTemplateBytes(template *template.Template, gameModeName string) []byte {
	var buffer bytes.Buffer
	handleFatalErr(template.Execute(&buffer, struct {
		GameModeName         string
		BackgroundImageUrl   redditImagePost
		NetworkPlayers       []onlinePlayer
		ServerDatas          map[string]*serverData
		NewPlayers           []newPlayer
		PotpissersTips       []string
		HcfTips              []string
		HcfClassTips         []string
		MzTips               []string
		Messages             []ingameMessage
		Announcements        []discordMessage
		Changelog            []discordMessage
		DiscordMessages      []discordMessage
		Donations            []order
		LineItemData         []lineItemData
		RedditVideos         []redditVideoPost
		DiscordId            string
		CurrentHcfServerName string
	}{
		GameModeName:         gameModeName,
		BackgroundImageUrl:   redditImagePosts[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(redditImagePosts))],
		NetworkPlayers:       currentPlayers,
		ServerDatas:          serverDatas,
		NewPlayers:           newPlayers,
		PotpissersTips:       potpissersTips,
		HcfTips:              cubecoreTips,
		HcfClassTips:         cubecoreClassTips,
		MzTips:               mzTips,
		Messages:             messages,
		Announcements:        announcements,
		Changelog:            changelog,
		DiscordMessages:      discordMessages,
		Donations:            donations,
		LineItemData:         lineItemDatas,
		RedditVideos:         redditVideoPosts,
		DiscordId:            "1245300045188956252",
		CurrentHcfServerName: currentHcfServerName,
	}))
	return buffer.Bytes()
}

var home []byte
var mz []byte
var hcf []byte

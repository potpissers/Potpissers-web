package main

import "github.com/jackc/pgx/v5"

var potpissersTips []string
var cubecoreTips []string
var cubecoreClassTips []string
var mzTips []string

func init() { // TODO -> move this to getting tips by name
	getRowsBlocking("SELECT * FROM get_tips()", func(rows pgx.Rows) {
		var tipMessage struct {
			gameModeName string
			tipTitle     string
			tipMessage   string
		}
		handleFatalPgx(pgx.ForEachRow(rows, []any{&tipMessage.gameModeName, &tipMessage.tipTitle, &tipMessage.tipMessage}, func() error {
			switch tipMessage.gameModeName {
			case "potpissers":
				potpissersTips = append(potpissersTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			case "cubecore":
				cubecoreTips = append(cubecoreTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			case "cubecore_classes":
				cubecoreClassTips = append(cubecoreClassTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			case "kollusion":
				mzTips = append(mzTips, tipMessage.tipTitle+": "+tipMessage.tipMessage)
			}
			return nil
		}))
	})
}
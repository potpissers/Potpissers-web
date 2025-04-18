package main

import "github.com/jackc/pgx/v5"

type gameModeTips struct {
	Title        string
	GameModeName string
}

var contentData = make(map[gameModeTips][]tip)

type tip struct {
	Title   string
	Message string
}

func init() { // TODO -> move this to getting tips by name
	getRowsBlocking("SELECT * FROM get_tips()", func(rows pgx.Rows) {
		var gameModeName string
		var tipTitle string
		var tipMessage string
		handleFatalPgx(pgx.ForEachRow(rows, []any{&gameModeName, &tipTitle, &tipMessage}, func() error {
			tip := tip{tipTitle, tipMessage}
			var gameModeTipsIteration gameModeTips
			switch gameModeName {
			case "potpissers":
				gameModeTipsIteration = gameModeTips{"potpissers tips", "hub"}
			case "potpissers-commands":
				gameModeTipsIteration = gameModeTips{"potpissers commands", ""}
			case "cubecore":
				gameModeTipsIteration = gameModeTips{"hcf tips", "hcf"}
			case "cubecore-commands":
				gameModeTipsIteration = gameModeTips{"hcf commands", ""}
			case "cubecore_classes":
				gameModeTipsIteration = gameModeTips{"hcf classes", ""}
			case "cubecore_classes_items":
				gameModeTipsIteration = gameModeTips{"hcf class items", ""}
			case "cubecore_recipes":
				gameModeTipsIteration = gameModeTips{"hcf recipes", ""}
			case "kollusion":
				gameModeTipsIteration = gameModeTips{"mz tips", "mz"}
			case "kollusion-commands":
				gameModeTipsIteration = gameModeTips{"mz commands", ""}
			}
			contentData[gameModeTipsIteration] = append(contentData[gameModeTipsIteration], tip)
			return nil
		}))
	})
	println("tips done")
}

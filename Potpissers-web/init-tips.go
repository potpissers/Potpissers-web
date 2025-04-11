package main

import "github.com/jackc/pgx/v5"

var contentData = make(map[string][]tip)

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
			switch gameModeName {
			case "potpissers":
				contentData["potpissers tips"] = append(contentData["potpissers tips"], tip)
			case "cubecore":
				contentData["hcf tips"] = append(contentData["hcf tips"], tip)
			case "cubecore_classes":
				contentData["hcf class tips"] = append(contentData["hcf class tips"], tip)
			case "kollusion":
				contentData["mz tips"] = append(contentData["mz tips"], tip)
			}
			return nil
		}))
	})
}

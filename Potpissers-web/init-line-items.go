package main

import "github.com/jackc/pgx/v5"

type lineItemData struct {
	GameModeName     string
	ItemName         string
	ItemPriceInCents int
	ItemDescription  string
	IsPlural         bool

	ItemPriceInDollars int
}

var lineItemDatas = func() []lineItemData {
	var slice []lineItemData
	getRowsBlocking("SELECT * FROM get_line_items()", func(rows pgx.Rows) {
		var death lineItemData
		handleFatalPgx(pgx.ForEachRow(rows, []any{&death.GameModeName, &death.ItemName, &death.ItemPriceInCents, &death.ItemDescription, &death.IsPlural}, func() error {
			death.ItemPriceInDollars = death.ItemPriceInCents / 100.0
			slice = append(slice, death)
			return nil
		}))
	})
	return slice
}()
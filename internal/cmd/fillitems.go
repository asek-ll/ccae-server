package cmd

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/asek-ll/aecc-server/internal/dao"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = FillItemsCommand{}

type FillItemsCommand struct {
}

func (s FillItemsCommand) Execute(args []string) error {

	db, err := sql.Open("sqlite3", "data.db")

	if err != nil {
		return err
	}

	_, err = db.Exec("DROP TABLE IF EXISTS item;")

	if err != nil {
		return err
	}

	items, err := dao.NewItemsDao(db)
	if err != nil {
		return err
	}

	jsonFile, err := os.Open("items.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data []map[string]any
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	for _, d := range data {
		var item dao.Item

		item.ID = d["id"].(string)
		item.DisplayName = d["displayName"].(string)
		if nbtRaw, ok := d["nbt"].(string); ok {
			item.NBT = &nbtRaw
		}
		if metaRaw, ok := d["meta"].(float64); ok {
			converted := int(metaRaw)
			item.Meta = &converted
		}
		iconEncoded := d["icon"].(string)

		item.Icon, err = base64.StdEncoding.DecodeString(iconEncoded)
		if err != nil {
			return err
		}

		err = items.InsertItems([]dao.Item{item})
		if err != nil {
			return err
		}
	}

	return nil
}

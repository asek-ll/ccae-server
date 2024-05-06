package cmd

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

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

	sqlStmt := `

	DROP TABLE IF EXISTS item;

	CREATE TABLE item (
		id string NOT NULL PRIMARY KEY,
		display_name string NOT NULL,
		nbt string,
		meta integer,
		icon BLOB
	);
	`
	_, err = db.Exec(sqlStmt)
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
		id := d["id"].(string)
		displayName := d["displayName"].(string)
		var nbt *string
		if nbtRaw, ok := d["nbt"].(string); ok {
			nbt = &nbtRaw
		}
		var meta *int
		if metaRaw, ok := d["meta"].(float64); ok {
			converted := int(metaRaw)
			meta = &converted
		}
		iconEncoded := d["icon"].(string)

		icon, err := base64.StdEncoding.DecodeString(iconEncoded)
		if err != nil {
			return err
		}

		_, err = db.Exec("INSERT INTO item VALUES(?, ?, ?, ?, ?)", id, displayName, nbt, meta, icon)
		if err != nil {
			return err
		}
	}

	return nil
}

package dao

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
)

type Item struct {
	UID         string
	ID          string
	DisplayName string
	NBT         *string
	Meta        *int
	Icon        []byte
}

func (i Item) Base64Icon() string {
	return base64.StdEncoding.EncodeToString(i.Icon)
}

type ItemsDao struct {
	db *sql.DB
}

func NewItemsDao(db *sql.DB) (*ItemsDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS item (
		uid string NOT NULL PRIMARY KEY,
		id string NOT NULL,
		display_name string NOT NULL,
		nbt string,
		meta integer,
		icon BLOB
	);
	`

	_, err := db.Exec(sqlStmt)

	if err != nil {
		return nil, err
	}

	return &ItemsDao{db: db}, nil
}

func (d *ItemsDao) InsertItems(items []Item) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO item (uid, id, display_name, nbt, meta, icon) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = stmt.Exec(common.MakeUid(item.ID, item.NBT), item.ID, item.DisplayName, item.NBT, item.Meta, item.Icon)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()

	return err
}

func readItemRows(rows *sql.Rows) ([]Item, error) {
	var result []Item
	for rows.Next() {
		var uid string
		var id string
		var displayName string
		var nbt *string
		var meta *int
		var icon []byte

		err := rows.Scan(&uid, &id, &displayName, &nbt, &meta, &icon)
		if err != nil {
			return nil, err
		}
		result = append(result, Item{
			UID:         uid,
			ID:          id,
			DisplayName: displayName,
			NBT:         nbt,
			Meta:        meta,
			Icon:        icon,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *ItemsDao) FindItemsByUids(uids []string) ([]Item, error) {
	if len(uids) == 0 {
		return nil, nil
	}

	rest := strings.Repeat(", ?", len(uids)-1)
	args := make([]any, len(uids))
	for i, uid := range uids {
		args[i] = uid
	}
	rows, err := d.db.Query(fmt.Sprintf("SELECT uid, id, display_name, nbt, meta, icon FROM item WHERE uid IN (?%s)", rest), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readItemRows(rows)
}

func (d *ItemsDao) FindByName(filter string) ([]Item, error) {

	rows, err := d.db.Query("SELECT uid, id, display_name, nbt, meta, icon FROM item WHERE display_name LIKE ? LIMIT 14", "%"+filter+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readItemRows(rows)
}

func (d *ItemsDao) FindItemByUid(uid string) (*Item, error) {
	items, err := d.FindItemsByUids([]string{uid})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("Item not found")
	}
	return &items[0], nil
}

func (d *ItemsDao) FindItemsIndexed(itemsByUid map[string]*Item) error {
	uids := common.MapKeys(itemsByUid)
	items, err := d.FindItemsByUids(uids)
	if err != nil {
		return err
	}

	for _, i := range items {
		itemsByUid[i.UID] = &i
	}

	return nil
}

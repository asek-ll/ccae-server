package dao

import (
	"database/sql"
	"fmt"
	"strings"
)

type Item struct {
	ID          string
	DisplayName string
	NBT         *string
	Meta        *int
	Icon        []byte
}

type ItemId struct {
	ID  string
	NBT string
}

func (i Item) UniqID() ItemId {
	var nbt string
	if i.NBT != nil {
		nbt = *i.NBT
	}
	return ItemId{
		ID:  i.ID,
		NBT: nbt,
	}
}

type ItemsDao struct {
	db *sql.DB
}

func NewItemsDao(db *sql.DB) *ItemsDao {
	return &ItemsDao{db: db}
}

func (d *ItemsDao) FindItemsByIds(ids []ItemId) ([]Item, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var parts []string
	var params []any

	for _, id := range ids {
		if id.NBT == "" {
			parts = append(parts, "id = ?")
			params = append(params, id.ID)
		} else {
			parts = append(parts, "id = ? AND nbt = ?")
			params = append(params, id.ID, id.NBT)
		}
	}

	conds := strings.Join(parts, ") OR (")

	fmt.Println(conds, params)

	rows, err := d.db.Query(fmt.Sprintf("SELECT * FROM item WHERE (%s)", conds), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Item
	for rows.Next() {
		var id string
		var displayName string
		var nbt *string
		var meta *int
		var icon []byte

		err = rows.Scan(&id, &displayName, &nbt, &meta, &icon)
		if err != nil {
			return nil, err
		}
		result = append(result, Item{
			ID:          id,
			DisplayName: displayName,
			NBT:         nbt,
			Meta:        meta,
			Icon:        icon,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

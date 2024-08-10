package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/asek-ll/aecc-server/internal/dao"
)

type Item struct {
	UID         string  `json:"uid"`
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	NBT         *string `json:"nbt,omitempty"`
	Meta        *int    `json:"meta,omitempty"`
	Icon        string  `json:"icon"`
}

func WriteJson(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	return encoder.Encode(data)
}

func ItemToDto(item *dao.Item) Item {
	return Item{
		UID:         item.UID,
		ID:          item.ID,
		DisplayName: item.DisplayName,
		NBT:         item.NBT,
		Meta:        item.Meta,
		Icon:        item.Base64Icon(),
	}
}

func ItemSuggest(itemDao *dao.ItemsDao) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		filter := r.URL.Query().Get("filter")
		items, err := itemDao.FindByName(filter)
		if err != nil {
			return err
		}
		resultItems := make([]Item, len(items))
		for i, item := range items {
			resultItems[i] = Item{
				UID:         item.UID,
				ID:          item.ID,
				DisplayName: item.DisplayName,
				NBT:         item.NBT,
				Meta:        item.Meta,
				Icon:        item.Base64Icon(),
			}
		}

		return WriteJson(w, resultItems)
	}
}

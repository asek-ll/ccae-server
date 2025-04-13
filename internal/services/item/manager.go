package item

import (
	"fmt"

	"github.com/asek-ll/aecc-server/internal/dao"
)

type ItemManager struct {
	daos *dao.DaoProvider
}

func NewItemManager(daos *dao.DaoProvider) *ItemManager {
	return &ItemManager{
		daos: daos,
	}
}

type ItemParams struct {
	NewUID      string
	DisplayName string
}

func (m *ItemManager) UpdateItem(uid string, params *ItemParams) (*dao.Item, error) {
	item, err := m.daos.Items.FindItemByUid(uid)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, fmt.Errorf("item with uid '%s' not found", uid)
	}

	if params.NewUID != "" {
		item.UID = params.NewUID
	}

	if params.DisplayName != "" {
		item.DisplayName = params.DisplayName
	}

	err = m.daos.Items.UpdateItem(uid, item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

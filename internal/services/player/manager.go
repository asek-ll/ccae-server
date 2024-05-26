package player

import (
	"context"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

type ItemStack struct {
	Name         string `json:"name"`
	Count        int    `json:"count"`
	MaxStackSize int    `json:"maxStackSize"`
	// Slot         int    `json:"slot"`
	// DisplayName  string `json:"displayName"`
	// tags: table 	A list of item tags
	// nbt: table 	The item's nbt data
}

type PlayerManager struct {
	ws          *wsrpc.JsonRpcServer
	daoProvider *dao.DaoProvider
}

func NewPlayerManager(ws *wsrpc.JsonRpcServer, daoProvider *dao.DaoProvider) *PlayerManager {
	return &PlayerManager{
		ws:          ws,
		daoProvider: daoProvider,
	}
}

func (s *PlayerManager) GetItems() (map[int]*crafter.Stack, error) {
	id, err := s.daoProvider.Clients.GetOnlineClientIdOfType("player")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var res map[int]ItemStack
	err = s.ws.SendRequestSync(ctx, id, "getItems", nil, &res)
	if err != nil {
		return nil, err
	}

	inventory := make(map[int]*crafter.Stack)
	for slot, item := range res {
		inventory[slot] = &crafter.Stack{
			ItemID: item.Name,
			Count:  item.Count,
		}

	}

	return inventory, nil
}

func (s *PlayerManager) RemoveItem(slot int) (int, error) {
	id, err := s.daoProvider.Clients.GetOnlineClientIdOfType("player")
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var res int
	err = s.ws.SendRequestSync(ctx, id, "removeItemFromPlayer", slot, &res)

	return res, err
}

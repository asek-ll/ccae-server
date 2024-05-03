package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

// type Storage interface {
// 	GetInventories() []Inventory
// 	GetItems() []StorageItem
// 	Transfer(from InventorySlot, to InventorySlot, amount int)
// }

type Storage struct {
	ws          *wsrpc.JsonRpcServer
	daoProvider *dao.DaoProvider
}

func NewStorage(ws *wsrpc.JsonRpcServer, daoProvider *dao.DaoProvider) *Storage {
	return &Storage{
		ws:          ws,
		daoProvider: daoProvider,
	}
}

func (s *Storage) GetItems() error {
	id, err := s.daoProvider.Clients.GetOnlineClientIdOfType("storage")
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	var res any
	err = s.ws.SendRequestSync(ctx, id, "getItems", nil, &res)
	if err != nil {
		return err
	}

	fmt.Println(res)

	return nil
}

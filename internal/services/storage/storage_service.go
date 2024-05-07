package storage

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

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

func keys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type AggregateStacks struct {
	Item  dao.Item
	Count int
	Image string
}

func (s *Storage) GetItems() ([]AggregateStacks, error) {
	id, err := s.daoProvider.Clients.GetOnlineClientIdOfType("storage")
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	var res []Inventory
	err = s.ws.SendRequestSync(ctx, id, "getItems", nil, &res)
	if err != nil {
		return nil, err
	}

	uniqueItems := make(map[string]Stack)

	for _, inv := range res {
		for _, item := range inv.Items {
			id := item.Item.GetUID()
			stack, e := uniqueItems[id]
			if !e {
				uniqueItems[id] = item.Item
			} else {
				stack.Count += item.Item.Count
				stack.MaxCount += item.Item.MaxCount
			}
		}
	}

	uids := keys(uniqueItems)

	items, err := s.daoProvider.Items.FindItemsByUids(uids)
	if err != nil {
		return nil, err
	}

	var stacks []AggregateStacks

	for _, item := range items {
		stacks = append(stacks, AggregateStacks{
			Item:  item,
			Count: uniqueItems[item.UID].Count,
			Image: base64.StdEncoding.EncodeToString(item.Icon),
		})
	}

	return stacks, nil
}

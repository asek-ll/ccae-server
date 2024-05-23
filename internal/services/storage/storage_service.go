package storage

import (
	"context"
	"errors"
	"time"

	"github.com/asek-ll/aecc-server/internal/common"
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

type AggregateStacks struct {
	Item  dao.Item
	Count int
}

func (s *Storage) GetItemsCount() (map[string]*Stack, error) {
	id, err := s.daoProvider.Clients.GetOnlineClientIdOfType("storage")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var res []Inventory
	err = s.ws.SendRequestSync(ctx, id, "getItems", nil, &res)
	if err != nil {
		return nil, err
	}

	uniqueItems := make(map[string]*Stack)

	for _, inv := range res {
		for _, item := range inv.Items {
			id := item.Item.GetUID()
			stack, e := uniqueItems[id]
			if !e {
				uniqueItems[id] = &item.Item
			} else {
				stack.Count += item.Item.Count
				stack.MaxCount += item.Item.MaxCount
			}
		}
	}

	return uniqueItems, nil
}

func (s *Storage) GetItems() ([]AggregateStacks, error) {
	uniqueItems, err := s.GetItemsCount()
	if err != nil {
		return nil, err
	}

	uids := common.MapKeys(uniqueItems)

	items, err := s.daoProvider.Items.FindItemsByUids(uids)
	if err != nil {
		return nil, err
	}

	var stacks []AggregateStacks

	for _, item := range items {
		stacks = append(stacks, AggregateStacks{
			Item:  item,
			Count: uniqueItems[item.UID].Count,
		})
	}

	return stacks, nil
}

type RichItemInfo struct {
	Item    dao.Item
	Recipes []*dao.Recipe
}

func (s *Storage) GetItemCount(uid string) (int, error) {
	return 0, nil
}

func (s *Storage) GetItem(uid string) (*RichItemInfo, error) {
	items, err := s.daoProvider.Items.FindItemsByUids([]string{uid})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("Item not found")
	}

	recipes, err := s.daoProvider.Recipes.GetRecipeByResult(uid)
	if err != nil {
		return nil, err
	}

	return &RichItemInfo{
		Item:    items[0],
		Recipes: recipes,
	}, nil
}

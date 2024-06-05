package storage

import (
	"errors"
	"log"
	"sort"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type Storage struct {
	daoProvider    *dao.DaoProvider
	clientsManager *wsmethods.ClientsManager
}

func NewStorage(daoProvider *dao.DaoProvider, clientsManager *wsmethods.ClientsManager) *Storage {
	return &Storage{
		daoProvider:    daoProvider,
		clientsManager: clientsManager,
	}
}

type AggregateStacks struct {
	Item  dao.Item
	Count int
}

func (s *Storage) GetItemsCount() (map[string]*wsmethods.Stack, error) {

	storageClient, err := wsmethods.GetClientForType[*wsmethods.StorageClient](s.clientsManager)
	if err != nil {
		return nil, err
	}

	res, err := storageClient.GetItems()
	if err != nil {
		return nil, err
	}

	uniqueItems := make(map[string]*wsmethods.Stack)

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
	log.Println("[INFO] GEt items")
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

	sort.Slice(stacks, func(a, b int) bool {
		return stacks[a].Count > stacks[b].Count
	})

	return stacks, nil
}

type RichItemInfo struct {
	Item    *dao.Item
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

	recipes, err := s.daoProvider.Recipes.GetRecipesByResults([]string{uid})
	if err != nil {
		return nil, err
	}

	return &RichItemInfo{
		Item:    &items[0],
		Recipes: recipes,
	}, nil
}

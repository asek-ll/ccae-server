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
	storageAdapter *wsmethods.StorageAdapter
	combinedStore  *CombinedStore
}

func NewStorage(daoProvider *dao.DaoProvider, clientsManager *wsmethods.ClientsManager) *Storage {
	adapter := wsmethods.NewStorageAdapter(clientsManager)
	return &Storage{
		daoProvider:    daoProvider,
		storageAdapter: adapter,
		combinedStore:  NewCombinedStore(adapter),
	}
}

type AggregateStacks struct {
	Item  dao.Item
	Count int
}

func (s *Storage) GetItemsCount() (map[string]*wsmethods.Stack, error) {

	res, err := s.storageAdapter.GetItems()

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

func (s *Storage) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	return s.combinedStore.ImportStack(uid, fromInventory, fromSlot, amount)
}

func (s *Storage) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	return s.combinedStore.ExportStack(uid, toInventory, toSlot, amount)
}

type ItemStore interface {
	ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error)
	ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error)
}

type SlotRef struct {
	Inventory string
	Slot      int
}

type IndexedInventory map[string]map[SlotRef]int

func indexInventory(invs []*wsmethods.Inventory) IndexedInventory {
	result := make(map[string]map[SlotRef]int)

	for _, inventory := range invs {
		slotRef := SlotRef{Inventory: inventory.Name}
		for _, stack := range inventory.Items {
			uid := stack.Item.GetUID()
			slotRef.Slot = stack.Slot
			stackMap, e := result[uid]
			if !e {
				stackMap = make(map[SlotRef]int)
				result[uid] = stackMap
			}
			stackMap[slotRef] = stack.Item.Count
		}
	}

	return result
}

package storage

import (
	"errors"
	"log"
	"sort"
	"sync"

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

	res, err := wsmethods.CallWithClientForType(s.clientsManager,
		func(client *wsmethods.StorageClient) ([]wsmethods.Inventory, error) {
			return client.GetItems()
		})

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

type MultipleChestsStore struct {
	StacksByUID    map[string][]StoreStack
	Inventories    map[string][]StoreInventory
	MaxSizeByUID   map[string]int
	mu             sync.RWMutex
	clientsManager *wsmethods.ClientsManager
}

func (s *MultipleChestsStore) GetMaxSize(UID string, inventory string, slot int) (int, error) {
	s.mu.RLock()
	size, e := s.MaxSizeByUID[UID]
	s.mu.RUnlock()
	if e {
		return size, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	size, e = s.MaxSizeByUID[UID]
	if e {
		return size, nil
	}

	res, err := wsmethods.CallWithClientForType(s.clientsManager,
		func(client *wsmethods.StorageClient) (*wsmethods.StackDetail, error) {
			return client.GetStackDetail(wsmethods.SlotRef{InventoryName: inventory, Slot: slot})
		})
	if err != nil {
		return 0, err
	}

	s.MaxSizeByUID[UID] = res.MaxCount
	return res.MaxCount, nil
}

func (s *MultipleChestsStore) ImportStack(UID string, slot int, inventory string, amount int) error {
	stacks := s.StacksByUID[UID]
	maxCount, err := s.GetMaxSize(UID, inventory, slot)
	if err != nil {
		return err
	}

	return nil
}

type ManagedStore struct {
	StacksByUID map[string]StoreStack
}

type StoreStack struct {
	Invenory string
	Slot     int
	Size     int
}
type StoreInventory struct {
	Name    string
	Stacks  map[int]wsmethods.Stack
	Size    int
	Managed bool
}
type Store struct {
	StacksByUID map[string][]StoreStack
	Inventories map[string][]StoreInventory
}

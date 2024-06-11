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

type ItemStore interface {
	ImportStack(uid string, slot int, inventory string, amount int) (int, error)
	ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error)
}

type MultipleChestsStore struct {
	StacksByUID    map[string][]StoreStack
	Inventories    map[string]StoreInventory
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

func (s *MultipleChestsStore) MoveStack(fromInventory string, fromSlot int, toInventory string, toSlot int, amount int) (int, error) {
	return wsmethods.CallWithClientForType(s.clientsManager, func(client *wsmethods.StorageClient) (int, error) {
		return client.MoveStack(fromInventory, fromSlot, toInventory, toSlot, amount)
	})
}

func (s *MultipleChestsStore) setStackSize(uid string, inventory string, slot int, amount int) {
	stacks := s.StacksByUID[uid]
	i := 0
	for ; i < len(stacks); i += 1 {
		stack := stacks[i]
		if stack.Inventory == inventory && stack.Slot == slot {
			break
		}
	}
	if amount == 0 {
		inv := s.Inventories[inventory]
		delete(inv.Stacks, slot)
		if i < len(stacks) {
			if len(stacks) == 0 {
				s.StacksByUID[uid] = nil
			} else {
				stacks[i] = stacks[len(stacks)-1]
				s.StacksByUID[uid] = stacks[0 : len(stacks)-1]
			}
		}
		return
	}
	s.Inventories[inventory].Stacks[slot] = Stack{
		UID:   uid,
		Count: amount,
	}

	stack := StoreStack{
		Inventory: inventory,
		Slot:      slot,
		Size:      amount,
	}

	if i < len(stacks) {
		stacks[i] = stack
	} else {
		s.StacksByUID[uid] = append(stacks, stack)
	}
}

func (s *MultipleChestsStore) ImportToEmptySlot(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	for name, inv := range s.Inventories {
		if inv.Size > len(inv.Stacks) {
			for i := 1; i <= inv.Size; i += 1 {
				if _, e := inv.Stacks[i]; !e {
					moved, err := s.MoveStack(fromInventory, fromSlot, name, i, amount)
					if err != nil {
						return 0, err
					}

					s.setStackSize(uid, name, i, amount)
					return moved, nil
				}
			}

		}
	}
	return 0, nil
}

func (s *MultipleChestsStore) ImportStack(uid string, slot int, inventory string, amount int) (int, error) {
	stacks := s.StacksByUID[uid]
	if len(stacks) == 0 {
		return s.ImportToEmptySlot(uid, inventory, slot, amount)
	}
	maxCount, err := s.GetMaxSize(uid, inventory, slot)
	if err != nil {
		return 0, err
	}

	remain := amount

	for _, stack := range stacks {
		toTransfer := min(maxCount-stack.Size, remain)
		if toTransfer > 0 {
			moved, err := s.MoveStack(inventory, slot, stack.Inventory, stack.Slot, toTransfer)
			if err != nil {
				return 0, err
			}
			s.setStackSize(uid, stack.Inventory, stack.Slot, stack.Size+moved)
			remain -= moved
			if remain == 0 {
				break
			}
		}
	}
	if remain > 0 {
		moved, err := s.ImportToEmptySlot(uid, inventory, slot, remain)
		if err != nil {
			return 0, err
		}
		remain -= moved
	}

	return amount - remain, nil
}

func (s *MultipleChestsStore) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	stacks := s.StacksByUID[uid]

	remain := amount
	for _, stack := range stacks {
		toTransfer := min(stack.Size, remain)
		if toTransfer > 0 {
			moved, err := s.MoveStack(stack.Inventory, stack.Slot, toInventory, toSlot, toTransfer)
			if err != nil {
				return 0, err
			}
			s.setStackSize(uid, stack.Inventory, stack.Slot, stack.Size-moved)
			remain -= moved
			if remain == 0 {
				break
			}
		}
	}

	return amount - remain, nil
}

type ManagedStore struct {
	StacksByUID map[string]StoreStack
}

type StoreStack struct {
	Inventory string
	Slot      int
	Size      int
}

type Stack struct {
	UID   string
	Count int
}

type StoreInventory struct {
	Name    string
	Stacks  map[int]Stack
	Size    int
	Managed bool
}

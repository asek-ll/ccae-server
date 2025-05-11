package storage

import (
	"errors"
	"log"
	"sort"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type ItemStore interface {
	ImportAll(fromInventory string) error
	ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error)
	ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error)
	GetItemsCount() (map[string]int, error)
}

type Storage struct {
	daoProvider    *dao.DaoProvider
	storageAdapter *wsmethods.StorageAdapter
	combinedStore  *CombinedStore
}

func NewStorage(daoProvider *dao.DaoProvider, storageAdapter *wsmethods.StorageAdapter) *Storage {
	return &Storage{
		daoProvider:    daoProvider,
		storageAdapter: storageAdapter,
		combinedStore:  NewCombinedStore(storageAdapter),
	}
}

type AggregateStacks struct {
	Item  dao.Item
	Count int
}

type StackGroup struct {
	Name   string
	Stacks []AggregateStacks
}

func (s *Storage) GetItemsCount() (map[string]int, error) {
	return s.combinedStore.GetItemsCount()
}

func (s *Storage) GetItems(filter string) ([]StackGroup, error) {
	log.Println("[INFO] Get items")
	groups, err := s.combinedStore.GetItemsGroupsCount()
	if err != nil {
		return nil, err
	}

	var uids []string
	visitedUids := make(map[string]struct{})
	for _, group := range groups {
		for key := range group.Counts {
			if _, e := visitedUids[key]; e {
				continue
			}
			uids = append(uids, key)
			visitedUids[key] = struct{}{}
		}
	}

	items, err := s.daoProvider.Items.FindItemsByUids(uids)
	if err != nil {
		return nil, err
	}
	itemsByUid := make(map[string]dao.Item)
	for _, item := range items {
		itemsByUid[item.UID] = item
	}

	filter = strings.ToLower(filter)

	var resultGroups []StackGroup

	for _, group := range groups {
		var stacks []AggregateStacks
		for uid, count := range group.Counts {
			item, ok := itemsByUid[uid]
			if !ok {
				item.ID = uid
				item.UID = uid
				item.DisplayName = uid
				item.Icon = common.QuestMarkIcon
			}
			if len(filter) == 0 || strings.Contains(strings.ToLower(item.DisplayName), filter) {
				stacks = append(stacks, AggregateStacks{
					Item:  item,
					Count: count,
				})
			}
		}
		sort.Slice(stacks, func(a, b int) bool {
			sa := stacks[a]
			sb := stacks[b]

			aisf := common.IsFluid(sa.Item.UID)
			bisf := common.IsFluid(sb.Item.UID)

			if aisf != bisf {
				return aisf
			}

			if sa.Count == sb.Count {
				return sa.Item.DisplayName < sb.Item.DisplayName
			}

			return stacks[a].Count > stacks[b].Count
		})
		if len(stacks) > 0 {
			resultGroups = append(resultGroups, StackGroup{
				Name:   group.Name,
				Stacks: stacks,
			})
		}
	}

	return resultGroups, nil
}

func (s *Storage) Optimize() error {
	return s.combinedStore.Optimize()
}

type RichItemInfo struct {
	Item            *dao.Item
	Recipes         []*dao.Recipe
	ImportedRecipes []*dao.Recipe
}

func (s *Storage) GetItemCount(uid string) (int, error) {
	counts, err := s.GetItemsCount()
	if err != nil {
		return 0, err
	}
	return counts[uid], nil
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

	importedRecipes, err := s.daoProvider.ImporetedRecipes.FindRecipeByResult(uid)
	if err != nil {
		return nil, err
	}
	if importedRecipes != nil {
		log.Printf("[WARN] Found imporeted recipe %v", *importedRecipes[0])
	} else {
		log.Printf("[WARN] No imporeted recipe found for %s", uid)
	}

	return &RichItemInfo{
		Item:            &items[0],
		Recipes:         recipes,
		ImportedRecipes: importedRecipes,
	}, nil
}

func (s *Storage) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	return s.combinedStore.ImportStack(uid, fromInventory, fromSlot, amount)
}

func (s *Storage) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	return s.combinedStore.ExportStack(uid, toInventory, toSlot, amount)
}

func (s *Storage) ImportFluid(uid string, fromInventory string, amount int) (int, error) {
	return s.combinedStore.fluidStorage.ImportFluid(uid, fromInventory, amount)
}

func (s *Storage) ExportFluid(uid string, toInventory string, amount int) (int, error) {
	return s.combinedStore.fluidStorage.ExportFluid(uid, toInventory, amount)
}

func (s *Storage) ImportAll(inventoryName string) error {
	items, err := s.storageAdapter.ListItems(inventoryName)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err := s.ImportStack(item.Item.GetUID(), inventoryName, item.Slot, item.Item.Count)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) ImportAllFluids(inventoryName string) error {
	tanks, err := s.storageAdapter.GetTanks(inventoryName)

	if err != nil {
		return err
	}

	for _, tank := range tanks {
		log.Printf("DUMP OUT %v", tank)
		_, err := s.ImportFluid(tank.Fluid.Name, inventoryName, tank.Fluid.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) ImportUnknownStack(inventoryName string, slot int) (int, error) {
	stack, err := s.storageAdapter.GetStackDetail(wsmethods.SlotRef{
		InventoryName: inventoryName,
		Slot:          slot,
	})
	if err != nil {
		return 0, err
	}

	return s.ImportStack(stack.GetUID(), inventoryName, slot, stack.Count)
}

func (s *Storage) GetInput() ([]string, error) {
	client, err := s.storageAdapter.GetClient()
	if err != nil {
		return nil, err
	}
	return client.InputStorages, nil
}

func (s *Storage) PullInputs() error {
	input, err := s.GetInput()
	if err != nil {
		return err
	}
	for _, inventoryName := range input {
		err = s.ImportAll(inventoryName)
		if err != nil {
			return err
		}
	}
	return nil
}

type SlotRef struct {
	Inventory string
	Slot      int
}

type Stack struct {
	UID   string
	Count int
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

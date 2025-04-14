package storage

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type CombinedStore struct {
	coldStorage  *SemiManagedStore
	warmStorage  *MultipleChestsStore
	fluidStorage *MultipleTanksStore

	loadedAt       time.Time
	storageAdapter *wsmethods.StorageAdapter

	mu sync.RWMutex
}

func NewCombinedStore(
	storageAdapter *wsmethods.StorageAdapter,
) *CombinedStore {
	store := &CombinedStore{
		coldStorage:    NewSemiManagedStore(storageAdapter),
		warmStorage:    NewMultipleChestsStore(storageAdapter),
		fluidStorage:   NewMultipleTanksStore(storageAdapter),
		storageAdapter: storageAdapter,
	}

	go store.syncCycle()

	return store
}

func (s *CombinedStore) syncCycle() {
	for {
		err := s.sync()
		if err != nil {
			time.Sleep(time.Second * 10)
		} else {
			time.Sleep(time.Second * 30)
		}
	}
}

func (s *CombinedStore) sync() error {
	client, err := s.storageAdapter.GetClient()
	if err != nil {
		return err
	}
	items, err := s.storageAdapter.GetItems([]string{client.ColdStoragePrefix, client.WarmStoragePrefix})
	if err != nil {
		return err
	}

	s.warmStorage.Clear()
	for _, inventory := range items {
		if strings.HasPrefix(inventory.Name, client.ColdStoragePrefix) {
			s.coldStorage.Sync(&inventory)
		} else if strings.HasPrefix(inventory.Name, client.WarmStoragePrefix) {
			s.warmStorage.Add(&inventory)
		} else {
			continue
		}
	}
	log.Println("[INFO] Items Synked!!!")

	containers, err := s.storageAdapter.GetFluidContainers([]string{client.SingleFluidContainerPrefix})
	if err != nil {
		return err
	}
	s.fluidStorage.Clear()
	for _, container := range containers {
		s.fluidStorage.Add(&container)
	}

	log.Println("[INFO] Fluid Synked!!!")
	s.loadedAt = time.Now()
	return nil
}

func (s *CombinedStore) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	movedCold, err := s.coldStorage.ImportStack(uid, fromInventory, fromSlot, amount)
	if err != nil {
		return 0, err
	}

	if movedCold < amount {
		movedWarm, err := s.warmStorage.ImportStack(uid, fromInventory, fromSlot, amount-movedCold)
		if err != nil {
			return 0, err
		}
		return movedCold + movedWarm, nil
	}

	return movedCold, nil
}

func (s *CombinedStore) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	movedCold, err := s.coldStorage.ExportStack(uid, toInventory, toSlot, amount)
	if err != nil {
		return 0, err
	}
	if movedCold < amount {
		movedWarm, err := s.warmStorage.ExportStack(uid, toInventory, toSlot, amount-movedCold)
		if err != nil {
			return 0, err
		}
		return movedCold + movedWarm, nil
	}

	return movedCold, nil
}

func (s *CombinedStore) GetItemsCount() (map[string]int, error) {
	count, err := s.coldStorage.GetItemsCount()
	if err != nil {
		return nil, err
	}
	warmCount, err := s.warmStorage.GetItemsCount()
	if err != nil {
		return nil, err
	}

	for k, v := range warmCount {
		count[k] += v
	}

	fluidsCount, err := s.fluidStorage.GetFluidsAmount()
	if err != nil {
		return nil, err
	}
	for k, v := range fluidsCount {
		count["fluid:"+k] += v
	}

	return count, nil
}

type ItemGroup struct {
	Name   string
	Counts map[string]int
}

func (s *CombinedStore) Optimize() error {
	coldStacks, err := s.coldStorage.GetStacks()
	if err != nil {
		return err
	}
	warmStacks, err := s.warmStorage.GetStacks()
	if err != nil {
		return err
	}

	for uid, stacks := range warmStacks {
		if _, e := coldStacks[uid]; e {
			for slot, count := range stacks {
				_, err := s.coldStorage.ImportStack(uid, slot.Inventory, slot.Slot, count)
				if err != nil {
					return err
				}
			}
		}
	}

	return s.sync()
}

func (s *CombinedStore) GetItemsGroupsCount() ([]ItemGroup, error) {
	var result []ItemGroup
	count, err := s.coldStorage.GetItemsCount()
	if err != nil {
		return nil, err
	}

	result = append(result, ItemGroup{
		Name:   "Cold Storage",
		Counts: count,
	})

	warmCount, err := s.warmStorage.GetItemsCount()
	if err != nil {
		return nil, err
	}

	result = append(result, ItemGroup{
		Name:   "Warm Storage",
		Counts: warmCount,
	})

	fluidsCount, err := s.fluidStorage.GetFluidsAmount()
	if err != nil {
		return nil, err
	}

	fixedFluids := make(map[string]int)
	for k, v := range fluidsCount {
		fixedFluids["fluid:"+k] = v
	}
	result = append(result, ItemGroup{
		Name:   "Fluid Storage",
		Counts: fixedFluids,
	})

	return result, nil
}

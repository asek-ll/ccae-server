package storage

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type CombinedStore struct {
	coldStorage    *SemiManagedStore
	warmStorage    *MultipleChestsStore
	loadedAt       time.Time
	storageAdapter *wsmethods.StorageAdapter

	mu sync.RWMutex
}

func NewCombinedStore(
	storageAdapter *wsmethods.StorageAdapter,
) *CombinedStore {
	return &CombinedStore{
		coldStorage:    NewSemiManagedStore(storageAdapter),
		warmStorage:    NewMultipleChestsStore(storageAdapter),
		storageAdapter: storageAdapter,
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
	log.Println("[INFO] Syned!!!")
	s.loadedAt = time.Now()
	return nil
}

func (s *CombinedStore) checkSync() error {
	if time.Since(s.loadedAt) > time.Second*30 {
		s.mu.Lock()
		defer s.mu.Unlock()

		if time.Since(s.loadedAt) > time.Second*30 {
			return s.sync()
		}
	}
	return nil
}

func (s *CombinedStore) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	err := s.checkSync()
	if err != nil {
		return 0, err
	}
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
	err := s.checkSync()
	if err != nil {
		return 0, err
	}
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
	err := s.checkSync()
	if err != nil {
		return nil, err
	}

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

	return count, nil
}

package storage

import (
	"strings"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type CombinedStore struct {
	coldStoragePrefix string
	warmStoragePrefix string

	coldStorage    *SemiManagedStore
	warmStorage    *MultipleChestsStore
	loadedAt       time.Time
	storageAdapter *wsmethods.StorageAdapter
	mu             sync.RWMutex
}

func NewCombinedStore(
	storageAdapter *wsmethods.StorageAdapter,
) *CombinedStore {
	return &CombinedStore{
		coldStoragePrefix: "storagedrowers:",
		warmStoragePrefix: "ironchest:diamond_chest",
		coldStorage:       NewSemiManagedStore(storageAdapter),
		warmStorage:       NewMultipleChestsStore(storageAdapter),
		storageAdapter:    storageAdapter,
	}
}

func (s *CombinedStore) sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.storageAdapter.GetItems()
	if err != nil {
		return err
	}

	s.warmStorage.Clear()
	for _, inventory := range items {
		if strings.HasPrefix(inventory.Name, s.coldStoragePrefix) {
			s.coldStorage.Sync(&inventory)
		}

		if strings.HasPrefix(inventory.Name, s.warmStoragePrefix) {
			s.warmStorage.Add(&inventory)
		}
	}
	s.loadedAt = time.Now()
	return nil
}

func (s *CombinedStore) checkSync() error {
	if time.Since(s.loadedAt) > time.Second*30 {
		return s.sync()
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

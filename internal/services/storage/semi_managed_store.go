package storage

import (
	"sync"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type SemiManagedStore struct {
	StacksByUID IndexedInventory
	itemStats   map[string]int

	storageAdapter *wsmethods.StorageAdapter
	mu             sync.RWMutex
}

func NewSemiManagedStore(storageAdapter *wsmethods.StorageAdapter) *SemiManagedStore {
	return &SemiManagedStore{
		StacksByUID:    make(IndexedInventory),
		itemStats:      make(map[string]int),
		storageAdapter: storageAdapter,
	}
}

func (s *SemiManagedStore) Sync(inventory *wsmethods.Inventory) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.StacksByUID = indexInventory([]*wsmethods.Inventory{inventory})
}

func (s *SemiManagedStore) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stacks := s.StacksByUID[uid]

	remain := amount
	for ref, count := range stacks {
		moved, err := s.storageAdapter.MoveStack(fromInventory, fromSlot, ref.Inventory, ref.Slot, remain)
		if err != nil {
			return 0, err
		}
		s.itemStats[uid] += moved
		stacks[ref] = count + moved
		remain -= moved
		if amount == 0 {
			break
		}
	}

	return amount - remain, nil
}

func (s *SemiManagedStore) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stacks := s.StacksByUID[uid]

	remain := amount
	for ref, count := range stacks {
		moved, err := s.storageAdapter.MoveStack(ref.Inventory, ref.Slot, toInventory, toSlot, remain)
		if err != nil {
			return 0, err
		}
		s.itemStats[uid] -= moved
		stacks[ref] = count - moved
		remain -= moved
		if amount == 0 {
			break
		}
	}

	return amount - remain, nil
}

func (s *SemiManagedStore) GetItemsCount() (map[string]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return common.CopyMap(s.itemStats), nil
}

package storage

import "github.com/asek-ll/aecc-server/internal/wsmethods"

type SemiManagedStore struct {
	StacksByUID    IndexedInventory
	storageAdapter *wsmethods.StorageAdapter
}

func NewSemiManagedStore(storageAdapter *wsmethods.StorageAdapter) *SemiManagedStore {
	return &SemiManagedStore{
		StacksByUID:    make(IndexedInventory),
		storageAdapter: storageAdapter,
	}
}

func (s *SemiManagedStore) Sync(inventory *wsmethods.Inventory) {
	s.StacksByUID = indexInventory([]*wsmethods.Inventory{inventory})
}

func (s *SemiManagedStore) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	stacks := s.StacksByUID[uid]

	remain := amount
	for ref, count := range stacks {
		moved, err := s.storageAdapter.MoveStack(fromInventory, fromSlot, ref.Inventory, ref.Slot, remain)
		if err != nil {
			return 0, err
		}
		stacks[ref] = count + moved
		remain -= moved
		if amount == 0 {
			break
		}
	}

	return amount - remain, nil
}

func (s *SemiManagedStore) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	stacks := s.StacksByUID[uid]

	remain := amount
	for ref, count := range stacks {
		moved, err := s.storageAdapter.MoveStack(ref.Inventory, ref.Slot, toInventory, toSlot, remain)
		if err != nil {
			return 0, err
		}
		stacks[ref] = count - moved
		remain -= moved
		if amount == 0 {
			break
		}
	}

	return amount - remain, nil
}
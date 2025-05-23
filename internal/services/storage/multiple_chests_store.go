package storage

import (
	"sync"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type StoreInventory struct {
	FreeSlots int
	Size      int
}

func NewMultipleChestsStore(storageAdapter *wsmethods.StorageAdapter) *MultipleChestsStore {
	return &MultipleChestsStore{
		stacksByUID:    make(map[string]map[SlotRef]int),
		inventories:    make(map[string]*StoreInventory),
		usedSlots:      make(map[SlotRef]struct{}),
		maxSizeByUID:   make(map[string]int),
		itemStats:      make(map[string]int),
		storageAdapter: storageAdapter,
	}
}

type MultipleChestsStore struct {
	stacksByUID  IndexedInventory
	inventories  map[string]*StoreInventory
	usedSlots    map[SlotRef]struct{}
	maxSizeByUID map[string]int
	itemStats    map[string]int

	storageAdapter *wsmethods.StorageAdapter
	mu             sync.RWMutex
}

func (s *MultipleChestsStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stacksByUID = make(map[string]map[SlotRef]int)
	s.usedSlots = make(map[SlotRef]struct{})
	s.inventories = make(map[string]*StoreInventory)
	s.itemStats = make(map[string]int)
}

func (s *MultipleChestsStore) Add(inv *wsmethods.Inventory) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inventories[inv.Name] = &StoreInventory{
		Size:      inv.Size,
		FreeSlots: inv.Size - len(inv.Items),
	}
	slotRef := SlotRef{Inventory: inv.Name}
	for _, stack := range inv.Items {
		uid := stack.Item.GetUID()
		slotRef.Slot = stack.Slot
		stackMap, e := s.stacksByUID[uid]
		if !e {
			stackMap = make(map[SlotRef]int)
			s.stacksByUID[uid] = stackMap
		}
		stackMap[slotRef] = stack.Item.Count
		s.usedSlots[slotRef] = struct{}{}
		s.itemStats[uid] += stack.Item.Count
	}
}

func (s *MultipleChestsStore) getMaxSize(UID string, inventory string, slot int) (int, error) {
	size, e := s.maxSizeByUID[UID]
	if e {
		return size, nil
	}

	size, e = s.maxSizeByUID[UID]
	if e {
		return size, nil
	}

	res, err := s.storageAdapter.GetStackDetail(wsmethods.SlotRef{InventoryName: inventory, Slot: slot})
	if err != nil {
		return 0, err
	}

	s.maxSizeByUID[UID] = res.MaxCount
	return res.MaxCount, nil
}

func (s *MultipleChestsStore) setStackSize(uid string, ref SlotRef, amount int) {
	stacks, e := s.stacksByUID[uid]
	if !e {
		stacks = make(map[SlotRef]int)
		s.stacksByUID[uid] = stacks
	}
	s.itemStats[uid] += amount - stacks[ref]
	if amount > 0 {
		if stacks[ref] == 0 {
			s.inventories[ref.Inventory].FreeSlots -= 1
			s.usedSlots[ref] = struct{}{}
		}
		stacks[ref] = amount
	} else {
		delete(stacks, ref)
		delete(s.usedSlots, ref)
		s.inventories[ref.Inventory].FreeSlots += 1
	}
}

func (s *MultipleChestsStore) importToEmptySlot(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	for name, inv := range s.inventories {
		if inv.FreeSlots == 0 {
			continue
		}
		ref := SlotRef{Inventory: name}
		for i := 1; i <= inv.Size; i += 1 {
			ref.Slot = i
			if _, e := s.usedSlots[ref]; !e {
				moved, err := s.storageAdapter.MoveStack(fromInventory, fromSlot, ref.Inventory, ref.Slot, amount)
				if err != nil {
					return 0, err
				}

				s.setStackSize(uid, ref, amount)
				return moved, nil
			}
		}

	}
	return 0, nil
}

func (s *MultipleChestsStore) ImportStack(uid string, fromInventory string, fromSlot int, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stacks := s.stacksByUID[uid]
	if len(stacks) == 0 {
		return s.importToEmptySlot(uid, fromInventory, fromSlot, amount)
	}
	maxCount, err := s.getMaxSize(uid, fromInventory, fromSlot)
	if err != nil {
		return 0, err
	}

	remain := amount

	for ref, count := range stacks {
		toTransfer := min(maxCount-count, remain)
		if toTransfer > 0 {
			moved, err := s.storageAdapter.MoveStack(fromInventory, fromSlot, ref.Inventory, ref.Slot, toTransfer)
			if err != nil {
				return 0, err
			}
			s.setStackSize(uid, ref, count+moved)
			remain -= moved
			if remain == 0 {
				break
			}
		}
	}
	if remain > 0 {
		moved, err := s.importToEmptySlot(uid, fromInventory, fromSlot, remain)
		if err != nil {
			return 0, err
		}
		remain -= moved
	}

	return amount - remain, nil
}

func (s *MultipleChestsStore) ExportStack(uid string, toInventory string, toSlot int, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stacks := s.stacksByUID[uid]

	remain := amount
	for ref, count := range stacks {
		toTransfer := min(count, remain)
		if toTransfer > 0 {
			moved, err := s.storageAdapter.MoveStack(ref.Inventory, ref.Slot, toInventory, toSlot, toTransfer)
			if err != nil {
				return 0, err
			}
			s.setStackSize(uid, ref, count-moved)
			remain -= moved
			if remain == 0 {
				break
			}
		}
	}

	return amount - remain, nil
}

func (s *MultipleChestsStore) GetItemsCount() (map[string]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return common.CopyMap(s.itemStats), nil
}

func (s *MultipleChestsStore) GetStacks() (IndexedInventory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return common.CopyMap(s.stacksByUID), nil
}

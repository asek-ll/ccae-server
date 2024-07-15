package storage

import (
	"sync"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type MultipleTanksStore struct {
	stacksByUID map[string]map[string]int
	itemStats   map[string]int
	emptyTanks  map[string]struct{}

	storageAdapter *wsmethods.StorageAdapter
	mu             sync.RWMutex
}

func NewMultipleTanksStore(storageAdapter *wsmethods.StorageAdapter) *MultipleTanksStore {
	return &MultipleTanksStore{
		stacksByUID:    make(map[string]map[string]int),
		itemStats:      make(map[string]int),
		emptyTanks:     make(map[string]struct{}),
		storageAdapter: storageAdapter,
	}
}

func (s *MultipleTanksStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stacksByUID = make(map[string]map[string]int)
	s.emptyTanks = make(map[string]struct{})
	s.itemStats = make(map[string]int)
}

func (s *MultipleTanksStore) Add(container *wsmethods.FluidContainer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(container.Tanks) == 0 {
		s.emptyTanks[container.Name] = struct{}{}
		return
	}

	for _, tank := range container.Tanks {
		fluidStack := tank.Fluid
		stackMap, e := s.stacksByUID[fluidStack.Name]
		if !e {
			stackMap = make(map[string]int)
			s.stacksByUID[fluidStack.Name] = stackMap
		}
		stackMap[container.Name] += fluidStack.Amount
		s.itemStats[fluidStack.Name] += fluidStack.Amount
	}
}

func (s *MultipleTanksStore) ImportFluid(uid string, fromContainer string, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stacks := s.stacksByUID[uid]
	if stacks == nil {
		if len(s.emptyTanks) == 0 {
			return 0, nil
		}
		stacks = make(map[string]int)
		for container := range s.emptyTanks {
			stacks[container] = 0
			delete(s.emptyTanks, container)
			break
		}
		s.stacksByUID[uid] = stacks
	}

	remain := amount
	for container, count := range stacks {
		moved, err := s.storageAdapter.MoveFluid(fromContainer, container, remain, uid)
		if err != nil {
			return 0, err
		}
		s.itemStats[uid] += moved
		stacks[container] = count + moved
		remain -= moved
		if amount == 0 {
			break
		}
	}

	return amount - remain, nil
}

func (s *MultipleTanksStore) ExportFluid(uid string, toContainer string, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stacks := s.stacksByUID[uid]

	remain := amount
	for container, count := range stacks {
		moved, err := s.storageAdapter.MoveFluid(container, toContainer, remain, uid)
		if err != nil {
			return 0, err
		}
		s.itemStats[uid] -= moved
		stacks[container] = count - moved
		remain -= moved
		if amount == 0 {
			break
		}
	}

	return amount - remain, nil
}

func (s *MultipleTanksStore) GetFluidsAmount() (map[string]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return common.CopyMap(s.itemStats), nil
}

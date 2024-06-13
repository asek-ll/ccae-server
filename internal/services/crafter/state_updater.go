package crafter

import (
	"sync"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type StateUpdater struct {
	storage storage.ItemStore
	daos    dao.DaoProvider
	crafter *Crafter

	snapshot map[string]int
	mu       sync.RWMutex
}

func (s *StateUpdater) UpdateState() error {
	counts, err := s.storage.GetItemsCount()
	if err != nil {
		return err
	}

	affectedPlanIds := make(map[int]struct{})

	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range counts {
		if v > s.snapshot[k] {
			planIds, err := s.daos.ItemReserves.UpdateItemCount(k, v)
			if err != nil {
				return err
			}
			for _, planId := range planIds {
				affectedPlanIds[planId] = struct{}{}
			}
		}
	}
	s.snapshot = counts
	s.mu.Unlock()

	for planId := range affectedPlanIds {
		state, err := s.daos.Plans.GetPlanById(planId)
		if err != nil {
			return err
		}
		err = s.crafter.CheckNextSteps(state)
		if err != nil {
			return err
		}
	}

	return nil
}

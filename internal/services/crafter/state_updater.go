package crafter

import (
	"log"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type StateUpdater struct {
	storage *storage.Storage
	daos    *dao.DaoProvider
	crafter *Crafter

	snapshot map[string]int
	mu       sync.RWMutex
}

func NewStateUpdater(
	storage *storage.Storage,
	daos *dao.DaoProvider,
	crafter *Crafter,
) *StateUpdater {
	return &StateUpdater{
		storage: storage,
		daos:    daos,
		crafter: crafter,
	}
}

func (s *StateUpdater) Start() {
	go func() {
		for {
			time.Sleep(time.Second * 30)
			err := s.UpdateState()
			if err != nil {
				log.Printf("[WARN] On state updater: %v", err)
			}
		}
	}()
}

func (s *StateUpdater) UpdateState() error {
	err := s.storage.PullInputs()
	if err != nil {
		return err
	}

	counts, err := s.storage.GetItemsCount()
	if err != nil {
		return err
	}

	affectedPlanIds := make(map[int]struct{})

	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range counts {
		if v > s.snapshot[k] {
			log.Printf("[INFO] Updated %s: %d > %d", k, v, s.snapshot[k])
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

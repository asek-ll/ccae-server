package player

import (
	"log"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type PlayerManager struct {
	daoProvider    *dao.DaoProvider
	clientsManager *wsmethods.ClientsManager
}

func NewPlayerManager(daoProvider *dao.DaoProvider, clientsManager *wsmethods.ClientsManager) *PlayerManager {
	pm := &PlayerManager{
		daoProvider:    daoProvider,
		clientsManager: clientsManager,
	}
	go func() {
		var slots []int
		for i := 18; i < 36; i += 1 {
			slots = append(slots, i)
		}
		for {
			time.Sleep(30 * time.Second)
			enabled, err := daoProvider.Configs.GetConfig("do-cleanup-inventory")
			log.Printf("[DEBUG] Read cleanup: %s", enabled)
			if err != nil {
				log.Println("[ERROR] Can't load config 'do-cleanup-inventory")
			} else if enabled == "true" {
				log.Println("[INFO] Do cleanup player iventory")
				pm.RemoveItems(slots)
			}
		}
	}()
	return pm
}

func (s *PlayerManager) GetItems() (map[int]*crafter.Stack, error) {
	playerClient, err := wsmethods.GetClientForType[*wsmethods.PlayerClient](s.clientsManager)
	if err != nil {
		return nil, err
	}

	res, err := playerClient.GetItems()
	if err != nil {
		return nil, err
	}

	inventory := make(map[int]*crafter.Stack)
	for slot, item := range res {
		inventory[slot] = &crafter.Stack{
			ItemID: item.Name,
			Count:  item.Count,
		}

	}

	return inventory, nil
}

func (s *PlayerManager) RemoveItem(slot int) (int, error) {
	playerClient, err := wsmethods.GetClientForType[*wsmethods.PlayerClient](s.clientsManager)
	if err != nil {
		return 0, err
	}

	return playerClient.RemoveItem(slot)
}

func (s *PlayerManager) RemoveItems(slots []int) error {
	playerClient, err := wsmethods.GetClientForType[*wsmethods.PlayerClient](s.clientsManager)
	if err != nil {
		return err
	}

	return playerClient.RemoveItems(slots)
}

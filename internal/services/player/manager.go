package player

import (
	"fmt"
	"log"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type PlayerManager struct {
	daoProvider    *dao.DaoProvider
	clientsManager *wsmethods.ClientsManager
	storage        *storage.Storage
}

func NewPlayerManager(daoProvider *dao.DaoProvider, clientsManager *wsmethods.ClientsManager, storage *storage.Storage) *PlayerManager {
	pm := &PlayerManager{
		daoProvider:    daoProvider,
		clientsManager: clientsManager,
		storage:        storage,
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

	err = playerClient.RemoveItems(slots)
	if err != nil {
		return err
	}

	bufferName, ok := playerClient.GetProps()["buffer_name"].(string)
	if !ok {
		return fmt.Errorf("invalid buffer_name: %v", playerClient.GetProps()["buffer_name"])
	}
	return s.storage.ImportAll(bufferName)
}

func (s *PlayerManager) SendItems(items []*crafter.Stack) error {
	playerClient, err := wsmethods.GetClientForType[*wsmethods.PlayerClient](s.clientsManager)
	if err != nil {
		return err
	}

	bufferName, ok := playerClient.GetProps()["buffer_name"].(string)
	if !ok {
		return fmt.Errorf("invalid buffer_name: %v", playerClient.GetProps()["buffer_name"])
	}
	log.Printf("[INFO] Send items to buffer: '%s'", bufferName)

	err = s.storage.ImportAll(bufferName)
	if err != nil {
		return err
	}

	slot := 1
	var targetSlots []int
	for _, item := range items {
		for item.Count > 0 {
			toTransfer := min(item.Count, 64)
			item.Count -= toTransfer
			_, err := s.storage.ExportStack(item.ItemID, bufferName, slot, toTransfer)
			if err != nil {
				return err
			}
			targetSlots = append(targetSlots, slot+17)
			slot += 1
		}
	}

	return playerClient.AddItems(targetSlots)
}

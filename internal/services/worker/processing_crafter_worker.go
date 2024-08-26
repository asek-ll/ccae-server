package worker

import (
	"fmt"
	"log"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/config"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type ProcessingCrafterWorker struct {
	daos           *dao.DaoProvider
	storage        *storage.Storage
	storageAdapter *wsmethods.StorageAdapter
	tm             *storage.TransferTransactionManager
}

func NewProcessingCrafterWorker(
	daos *dao.DaoProvider,
	storage *storage.Storage,
	storageAdapter *wsmethods.StorageAdapter,
	tm *storage.TransferTransactionManager,
) *ProcessingCrafterWorker {
	return &ProcessingCrafterWorker{
		daos:           daos,
		storage:        storage,
		storageAdapter: storageAdapter,
		tm:             tm,
	}
}

func isMatch[T comparable](values []T, value T) bool {
	if len(values) == 0 {
		return true
	}
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func (w *ProcessingCrafterWorker) pullItemsResults(config config.ProcessCrafterConfig) error {
	if config.ResultInventory == "" {
		return nil
	}
	items, err := w.storageAdapter.ListItems(config.ResultInventory)

	if err != nil {
		return err
	}
	for _, item := range items {
		uid := item.Item.GetUID()
		if isMatch(config.ResultInventorySlots, item.Slot) && isMatch(config.ResultItems, uid) {
			_, err := w.storage.ImportStack(uid, config.ResultInventory, item.Slot, item.Item.Count)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *ProcessingCrafterWorker) pullFluidResults(config config.ProcessCrafterConfig) error {
	if config.ResultTank == "" {
		return nil
	}
	fluids, err := w.storageAdapter.GetTanks(config.ResultTank)

	if err != nil {
		return err
	}
	for _, fluid := range fluids {
		uid := fluid.Fluid.Name

		if isMatch(config.ResultFluids, uid) {
			_, err := w.storage.ImportFluid(uid, config.ResultTank, fluid.Fluid.Amount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *ProcessingCrafterWorker) canProcess(config config.ProcessCrafterConfig, recipe *dao.Recipe) (bool, error) {
	if config.ReagentMode == "block" {
		if config.InputInventory != "" {
			items, err := w.storageAdapter.ListItems(config.InputInventory)
			if err != nil {
				return false, err
			}
			if len(items) > 0 {
				return false, nil
			}
		}
		if config.InputTank != "" {
			fluids, err := w.storageAdapter.GetTanks(config.InputTank)
			if err != nil {
				return false, err
			}
			if len(fluids) > 0 {
				return false, nil
			}
		}
		return true, nil
	}
	if config.ReagentMode == "same" {
		recipeFluids := make(map[string]struct{})
		recipeItems := make(map[string]struct{})
		for _, ing := range recipe.Ingredients {
			if common.IsFluid(ing.ItemUID) {
				recipeFluids[ing.ItemUID] = struct{}{}
			} else {
				recipeItems[ing.ItemUID] = struct{}{}
			}
		}
		if config.InputInventory != "" && len(recipeItems) > 0 {
			items, err := w.storageAdapter.ListItems(config.InputInventory)
			if err != nil {
				return false, err
			}
			for _, item := range items {
				if _, e := recipeItems[item.Item.GetUID()]; !e {
					return false, nil
				}
			}
		}
		if config.InputTank != "" && len(recipeFluids) > 0 {
			tanks, err := w.storageAdapter.GetTanks(config.InputTank)
			if err != nil {
				return false, err
			}
			for _, fluid := range tanks {
				if _, e := recipeFluids[fluid.Fluid.Name]; !e {
					return false, nil
				}
			}
		}
	}
	return true, nil
}

func (w *ProcessingCrafterWorker) do(config config.ProcessCrafterConfig) error {

	checkNext := true

	var err error

	err = w.pullItemsResults(config)
	if err != nil {
		log.Printf("[WARN] Error on pull items: %v", err)
	}

	err = w.pullFluidResults(config)
	if err != nil {
		log.Printf("[WARN] Error on pull fluids: %v", err)
	}

	for checkNext {
		checkNext = false
		crafts, err := w.daos.Crafts.FindNextByTypes([]string{config.CraftType}, "unknown")
		if err != nil {
			return err
		}
		for _, craft := range crafts {

			if config.ReagentMode == "block" {
				if config.InputInventory != "" {
					items, err := w.storageAdapter.ListItems(config.InputInventory)
					if err != nil {
						return err
					}
					if len(items) > 0 {
						return nil
					}
				}
				if config.InputTank != "" {
					fluids, err := w.storageAdapter.GetTanks(config.InputTank)
					if err != nil {
						return err
					}
					if len(fluids) > 0 {
						return nil
					}
				}
			}

			recipe, err := w.daos.Recipes.GetRecipeById(craft.RecipeID)
			if err != nil {
				return err
			}

			var req storage.ExportRequest
			slot := 0
			for _, ing := range recipe.Ingredients {
				if common.IsFluid(ing.ItemUID) {
					if config.InputTank == "" {
						return fmt.Errorf("Input tank not set")
					}
					req.RequestFluids = append(req.RequestFluids, storage.ExportRequestFluids{
						TargetTankName: config.InputTank,
						Uid:            common.FluidUid(ing.ItemUID),
						Amount:         ing.Amount,
					})
				} else {
					if config.InputInventory == "" {
						return fmt.Errorf("Input storage not set")
					}
					slot += 1
					if ing.Slot != nil {
						slot = *ing.Slot
					}
					req.RequestItems = append(req.RequestItems, storage.ExportRequestItems{
						TargetStorage: config.InputInventory,
						Uid:           ing.ItemUID,
						ToSlot:        slot,
						Amount:        ing.Amount,
					})
				}
			}

			tx, err := w.tm.CreateExportTransaction(req)
			if err != nil {
				return err
			}

			defer tx.Rollback()

			err = w.daos.Crafts.CommitCraftInOuterTx(tx.DBTx, craft, recipe, 1)
			if err != nil {
				return err
			}

			err = w.daos.Crafts.CompleteCraftInOuterTx(tx.DBTx, craft)

			if err != nil {
				return err
			}

			err = tx.Commit()
			if err != nil {
				return err
			}

			checkNext = true
		}
	}

	return nil
}

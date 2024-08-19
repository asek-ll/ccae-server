package worker

import (
	"fmt"

	"github.com/asek-ll/aecc-server/internal/common"
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

func (w *ProcessingCrafterWorker) do(config *dao.ProcessingCrafterWorkerConfig) error {

	checkNext := true

	for checkNext {
		checkNext = false
		crafts, err := w.daos.Crafts.FindNextByTypes([]string{config.CraftType}, "unknown")
		if err != nil {
			return err
		}
		for _, craft := range crafts {

			if config.ReagentMode == "block" {
				if config.InputStorage != "" {
					items, err := w.storageAdapter.ListItems(config.InputStorage)
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
					if config.InputStorage == "" {
						return fmt.Errorf("Input storage not set")
					}
					slot += 1
					if ing.Slot != nil {
						slot = *ing.Slot
					}
					req.RequestItems = append(req.RequestItems, storage.ExportRequestItems{
						TargetStorage: config.InputStorage,
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

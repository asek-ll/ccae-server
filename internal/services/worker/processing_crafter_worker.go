package worker

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type ProcessingCrafterWorker struct {
	daos    *dao.DaoProvider
	storage storage.Storage
	tm      *storage.TransferTransactionManager
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

			recipe, err := c.daos.Recipes.GetRecipeById(craft.RecipeID)
			if err != nil {
				return err
			}

			tx, err := w.tm.CreateExportTransaction(craft.)
			if err != nil {
				return err
			}

			defer tx.Rollback()

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

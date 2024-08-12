package worker

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type ProcessingCrafterWorker struct {
	daos    *dao.DaoProvider
	storage storage.Storage
}

func (w *ProcessingCrafterWorker) do(config *dao.ProcessingCrafterWorkerConfig) error {
	current, err := w.daos.Crafts.FindCurrent(config.CraftType)

	return nil
}

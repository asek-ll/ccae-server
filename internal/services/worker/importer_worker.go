package worker

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type ImporterWorker struct {
	storage storage.Storage
}

func NewImporterWorker(storage storage.Storage) *ImporterWorker {
	return &ImporterWorker{
		storage: storage,
	}
}

func (w *ImporterWorker) do(config *dao.ImporterWorkerConfig) error {

	for _, importConfig := range config.Imports {
		if importConfig.Slot == 0 {
			err := w.storage.ImportAll(importConfig.Storage)
			if err != nil {
				return err
			}
		} else {
			_, err := w.storage.ImportUnknownStack(importConfig.Storage, importConfig.Slot)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

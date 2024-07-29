package worker

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type ExporterWorker struct {
	storage storage.Storage
}

func NewExporterWorker(storage storage.Storage) *ExporterWorker {
	return &ExporterWorker{
		storage: storage,
	}
}

func (w *ExporterWorker) do(config *dao.ExporterWorkerConfig) error {

	for _, exportConfig := range config.Exports {
		_, err := w.storage.ExportStack(exportConfig.Storage, exportConfig.Storage, exportConfig.Slot, exportConfig.Amount)
		if err != nil {
			return err
		}

	}

	return nil
}

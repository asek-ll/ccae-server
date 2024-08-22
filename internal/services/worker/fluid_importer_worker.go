package worker

import (
	"log"

	"github.com/asek-ll/aecc-server/internal/config"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type FluidImporterWorker struct {
	storage storage.Storage
	configs []config.FluidImporterConfig
}

func NewFluidImporterWorker(
	storage storage.Storage,
	configs []config.FluidImporterConfig,
) *FluidImporterWorker {
	return &FluidImporterWorker{
		storage: storage,
		configs: configs,
	}
}

func (w *FluidImporterWorker) do() error {

	for _, importConfig := range w.configs {
		for _, fluid := range importConfig.Fluids {
			log.Println("IMPORT", fluid)
			_, err := w.storage.ImportFluid(fluid, importConfig.Tank, 1000)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

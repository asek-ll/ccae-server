package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/asek-ll/aecc-server/internal/config"
	"github.com/asek-ll/aecc-server/internal/dao"
)

type WorkerManager struct {
	workerHandlers       *WorkerHandlerManager
	daos                 *dao.DaoProvider
	exporterWorker       *ExporterWorker
	importerWorker       *ImporterWorker
	processCrafterWorker *ProcessingCrafterWorker
	configLoader         *config.ConfigLoader
	fluidImporterWorker  *FluidImporterWorker
}

func NewWorkerManager(
	config *config.ConfigLoader,
	daos *dao.DaoProvider,
	exporterWorker *ExporterWorker,
	importerWorker *ImporterWorker,
	processCrafterWorker *ProcessingCrafterWorker,
	fluidImporterWorker *FluidImporterWorker,
) *WorkerManager {
	wm := &WorkerManager{
		workerHandlers:       NewWorkerHandlerManager(),
		configLoader:         config,
		daos:                 daos,
		exporterWorker:       exporterWorker,
		importerWorker:       importerWorker,
		processCrafterWorker: processCrafterWorker,
		fluidImporterWorker:  fluidImporterWorker,
	}
	wm.init()

	return wm
}

func (w *WorkerManager) getRunner(worker *dao.Worker) func() error {
	switch worker.Type {
	case dao.WORKER_TYPE_EXPORTER:
		return func() error {
			log.Printf("%s worker tick", worker.Key)
			return w.exporterWorker.do(worker.Config.Exporter)
		}
	case dao.WORKER_TYPE_IMPORTER:
		return func() error {
			log.Printf("%s worker tick", worker.Key)
			return w.importerWorker.do(worker.Config.Importer)
		}
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		cfg := worker.Config.ProcessingCrafter
		workerConfig := config.ProcessCrafterConfig{
			WorkerKey:      worker.Key,
			CraftType:      cfg.CraftType,
			InputInventory: cfg.InputInventory,
			InputTank:      cfg.InputTank,
			ReagentMode:    cfg.ReagentMode,
			Enabled:        true,
			WaitResults:    cfg.WaitResults && worker.Key != "",
			CraftCondition: cfg.CraftCondition,

			ResultItems:          cfg.ResultItems,
			ResultInventory:      cfg.ResultInventory,
			ResultInventorySlots: cfg.ResultInventorySlots,

			ResultFluids: cfg.ResultFluids,
			ResultTank:   cfg.ResultTank,
		}
		return func() error {
			log.Printf("%s processer crafter tick tick", worker.Key)
			return w.processCrafterWorker.do(workerConfig)
		}
	}
	return func() error {
		log.Printf("NOP %s worker tick", worker.Key)
		return nil
	}
}

func (w *WorkerManager) addAndStart(worker *dao.Worker) error {
	handler, err := w.workerHandlers.Add(worker.Key, w.getRunner(worker))
	if err != nil {
		return err
	}
	handler.Start()
	return nil
}

func (w *WorkerManager) init() error {
	workers, err := w.daos.Workers.GetWorkers()
	if err != nil {
		return err
	}

	for _, worker := range workers {
		if worker.Enabled {
			err := w.addAndStart(&worker)
			if err != nil {
				return err
			}
		}
	}

	for _, pc := range w.configLoader.Config.Crafters.ProcessCrafters {
		if pc.Enabled {
			handler, err := w.workerHandlers.Add(fmt.Sprintf("pc_%s", pc.CraftType), func() error {
				return w.processCrafterWorker.do(pc)
			})
			if err != nil {
				return err
			}
			err = handler.Start()
			if err != nil {
				return err
			}
		}
	}

	if len(w.configLoader.Config.Importers.FluidImporters) > 0 {
		handler, err := w.workerHandlers.Add("fluidimporter", func() error {
			return w.fluidImporterWorker.do()
		})
		if err != nil {
			return err
		}
		err = handler.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *WorkerManager) CreateWorker(worker *dao.Worker) error {
	err := w.daos.Workers.CreateWorker(worker)
	if err != nil {
		return err
	}
	return w.addAndStart(worker)
}

func (w *WorkerManager) UpdateWorker(key string, worker *dao.Worker) error {
	err := w.daos.Workers.UpdateWorker(key, worker)
	if err != nil {
		return err
	}

	handler, err := w.workerHandlers.Add(worker.Key, w.getRunner(worker))
	if err != nil {
		return err
	}
	if worker.Enabled {
		handler.Start()
	}
	return nil
}

func (w *WorkerManager) DeleteWorker(key string) error {
	w.workerHandlers.Remove(key)

	err := w.daos.Workers.DeleteWorker(key)
	if err != nil {
		return err
	}
	return nil
}

func (w *WorkerManager) GetWorkers() ([]dao.Worker, error) {
	return w.daos.Workers.GetWorkers()
}

func (w *WorkerManager) GetWorker(key string) (*dao.Worker, error) {
	return w.daos.Workers.GetWorker(key)
}

func (w *WorkerManager) ParseWorker(params *WorkerParams) (*dao.Worker, error) {
	if params.Key == "" {
		return nil, errors.New("worker key is required")
	}

	if params.Type == "" {
		return nil, errors.New("worker type is required")
	}

	config := dao.WorkerConfig{}

	var err error
	switch params.Type {
	case dao.WORKER_TYPE_EXPORTER:
		config.Exporter, err = parseExporterWorkerConfig(params.Config.Exporter)
	case dao.WORKER_TYPE_IMPORTER:
		config.Importer, err = parseImporterWorkerConfig(params.Config.Importer)
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		config.ProcessingCrafter, err = w.parseProcessingCrafterWorkerConfig(params.Config.ProcessingCrafter)
	default:
		return nil, errors.New("invalid worker type")
	}
	if err != nil {
		return nil, err
	}

	return &dao.Worker{
		Key:     params.Key,
		Type:    params.Type,
		Enabled: params.Enabled,
		Config:  config,
	}, nil
}

func parseExporterWorkerConfig(params *ExporterWorkerConfigParams) (*dao.ExporterWorkerConfig, error) {
	config := dao.ExporterWorkerConfig{}
	for _, exportConfig := range params.Exports {

		if exportConfig.Storage == "" {
			return nil, fmt.Errorf("empty storage config")
		}

		if exportConfig.Item == "" {
			return nil, fmt.Errorf("empty item config")
		}

		slot, err := strconv.Atoi(exportConfig.Slot)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.Atoi(exportConfig.Amount)
		if err != nil {
			return nil, err
		}

		config.Exports = append(config.Exports, dao.SingleExportConfig{
			Storage: exportConfig.Storage,
			Item:    exportConfig.Item,
			Slot:    slot,
			Amount:  amount,
		})
	}
	if len(config.Exports) == 0 {
		return nil, fmt.Errorf("empty export configs")
	}
	return &config, nil
}

func parseImporterWorkerConfig(params *ImporterWorkerConfigParams) (*dao.ImporterWorkerConfig, error) {
	config := dao.ImporterWorkerConfig{}

	for _, importConfig := range params.Imports {

		if importConfig.Storage == "" {
			return nil, fmt.Errorf("empty storage config")
		}
		slot, err := strconv.Atoi(importConfig.Slot)
		if err != nil {
			return nil, err
		}

		config.Imports = append(config.Imports, dao.SingleImportConfig{
			Storage: importConfig.Storage,
			Slot:    slot,
		})
	}
	if len(config.Imports) == 0 {
		return nil, fmt.Errorf("empty imports configs")
	}
	return &config, nil

}
func (w *WorkerManager) parseProcessingCrafterWorkerConfig(params *ProcessingCrafterWorkerConfigParams) (*dao.ProcessingCrafterWorkerConfig, error) {
	config := dao.ProcessingCrafterWorkerConfig{}

	if params == nil || params.RawConfig == "" {
		return nil, errors.New("empty processing crafter config")
	}

	err := json.Unmarshal([]byte(params.RawConfig), &config)

	if err != nil {
		return nil, fmt.Errorf("processing crafter config parse error: %v", err)
	}

	config.CraftType = params.CraftType

	recipeType, err := w.daos.RecipeTypes.GetRecipeType(config.CraftType)
	if err != nil {
		return nil, fmt.Errorf("recipe type get: %v", err)
	}

	if recipeType == nil {
		return nil, fmt.Errorf("recipe type not found")
	}

	return &config, nil
}

func (w *WorkerManager) WorkerToParams(worker *dao.Worker) *WorkerParams {
	return NewWorkerParams(worker)
}

func (w *WorkerManager) AddWorkerItemUids(params *WorkerParams, itemLoader *dao.ItemDeferedLoader) {
	if params.Config.Exporter != nil {
		for _, exportConfig := range params.Config.Exporter.Exports {
			itemLoader.AddUid(exportConfig.Item)
		}
	}
}

func (w *WorkerManager) ParseWorkerParams(values url.Values) *WorkerParams {
	return ParseWorkerParams(values)
}

func (w *WorkerManager) NewWorkerParamsForType(workerType string) *WorkerParams {
	config := WorkerConfigParams{}
	switch workerType {
	case dao.WORKER_TYPE_EXPORTER:
		config.Exporter = &ExporterWorkerConfigParams{
			Exports: make([]SingleExportConfigParams, 1),
		}
	case dao.WORKER_TYPE_IMPORTER:
		config.Importer = &ImporterWorkerConfigParams{
			Imports: make([]SingleImportConfigParams, 1),
		}
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		config.ProcessingCrafter = &ProcessingCrafterWorkerConfigParams{
			CraftType: "",
			RawConfig: "",
		}
	}

	return &WorkerParams{
		Key:     "",
		Type:    workerType,
		Enabled: true,
		Config:  config,
	}
}

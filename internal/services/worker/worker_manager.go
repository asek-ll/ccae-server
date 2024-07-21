package worker

import (
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/asek-ll/aecc-server/internal/dao"
)

type WorkerManager struct {
	workerHandlers *WorkerHandlerManager
	daos           *dao.DaoProvider
}

func NewWorkerManager(daos *dao.DaoProvider) *WorkerManager {
	wm := &WorkerManager{
		workerHandlers: NewWorkerHandlerManager(),
		daos:           daos,
	}
	wm.init()
	return wm
}

func (w *WorkerManager) getRunner(worker *dao.Worker) func() error {
	// switch worker.Type {
	// case dao.WORKER_TYPE_SHAPED_CRAFTER:
	// 	return func() error {
	// 		return w.craftWorker(worker)
	// 	}
	// }
	return func() error {
		log.Printf("%s worker tick", worker.Key)
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
	return w.addAndStart(worker)
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

func (w *WorkerManager) ParseWorker(values url.Values) (*dao.Worker, error) {
	key := values.Get("key")
	if key == "" {
		return nil, errors.New("worker key is required")
	}

	workerType := values.Get("type")
	if workerType == "" {
		return nil, errors.New("worker type is required")
	}

	config := dao.WorkerConfig{}

	var err error
	switch workerType {
	case dao.WORKER_TYPE_EXPORTER:
		config.Exporter, err = parseExporterWorkerConfig(values)
	case dao.WORKER_TYPE_IMPORTER:
		config.Importer, err = parseImporterWorkerConfig(values)
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		config.ProcessingCrafter, err = parseProcessingCrafterWorkerConfig(values)
	default:
		return nil, errors.New("invalid worker type")
	}
	if err != nil {
		return nil, err
	}

	return &dao.Worker{
		Key:     key,
		Type:    workerType,
		Enabled: true,
		Config:  config,
	}, nil
}

func parseExporterWorkerConfig(values url.Values) (*dao.ExporterWorkerConfig, error) {
	config := dao.ExporterWorkerConfig{}

	storages := strings.Split(values.Get("storages"), ",")
	if len(storages) == 0 {
		return nil, errors.New("exporter storages is required")
	}
	config.Storages = storages

	items := strings.Split(values.Get("items"), ",")
	if len(items) == 0 {
		return nil, errors.New("exporter items is required")
	}
	config.Items = items

	return &config, nil
}
func parseImporterWorkerConfig(values url.Values) (*dao.ImporterWorkerConfig, error) {
	config := dao.ImporterWorkerConfig{}

	storages := strings.Split(values.Get("storages"), ",")
	if len(storages) == 0 {
		return nil, errors.New("exporter storages is required")
	}
	config.Storages = storages
	return &config, nil
}
func parseProcessingCrafterWorkerConfig(values url.Values) (*dao.ProcessingCrafterWorkerConfig, error) {
	config := dao.ProcessingCrafterWorkerConfig{}

	craftType := values.Get("craftType")
	if craftType == "" {
		return nil, errors.New("processing crafter craft type is required")
	}
	config.CraftType = craftType

	inputStorage := values.Get("inputStorage")
	if inputStorage == "" {
		return nil, errors.New("processing crafter input storage is required")
	}
	config.InputStorage = inputStorage

	reagentMode := values.Get("reagentMode")
	if reagentMode == "" {
		return nil, errors.New("processing crafter reagent mode is required")
	}
	config.ReagentMode = reagentMode

	return &config, nil
}

func (w *WorkerManager) WorkerToParams(worker *dao.Worker) (url.Values, error) {
	params := make(url.Values)
	params.Set("key", worker.Key)
	params.Set("type", worker.Type)

	switch worker.Type {
	case dao.WORKER_TYPE_EXPORTER:
		params.Set("storages", strings.Join(worker.Config.Exporter.Storages, ","))
		params.Set("items", strings.Join(worker.Config.Exporter.Items, ","))
	case dao.WORKER_TYPE_IMPORTER:
		params.Set("storages", strings.Join(worker.Config.Importer.Storages, ","))
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		params.Set("inputStorage", worker.Config.ProcessingCrafter.InputStorage)
		params.Set("craftType", worker.Config.ProcessingCrafter.CraftType)
		params.Set("reagentMode", worker.Config.ProcessingCrafter.ReagentMode)
	default:
		return nil, errors.New("invalid worker type")
	}

	return params, nil
}

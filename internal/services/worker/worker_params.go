package worker

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
)

type SingleExportConfigParams struct {
	Storage string
	Item    string
	Slot    string
	Amount  string
}

type ExporterWorkerConfigParams struct {
	Exports []SingleExportConfigParams
}

type SingleImportConfigParams struct {
	Storage string
	Slot    string
}

type ImporterWorkerConfigParams struct {
	Imports []SingleImportConfigParams
}

type WorkerConfigParams struct {
	Exporter          *ExporterWorkerConfigParams
	Importer          *ImporterWorkerConfigParams
	ProcessingCrafter *dao.ProcessingCrafterWorkerConfig
}

type WorkerParams struct {
	Key     string
	Type    string
	Enabled bool
	Config  WorkerConfigParams
}

func ParseWorkerParams(values url.Values) *WorkerParams {
	key := values.Get("key")
	workerType := values.Get("type")

	config := WorkerConfigParams{}
	switch workerType {
	case dao.WORKER_TYPE_EXPORTER:
		config.Exporter = parseExporterWorkerConfigParams(values)
	case dao.WORKER_TYPE_IMPORTER:
		config.Importer = parseImporterWorkerConfigParams(values)
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		config.ProcessingCrafter = parseProcessingCrafterWorkerConfigParams(values)
	}

	return &WorkerParams{
		Key:     key,
		Type:    workerType,
		Enabled: true,
		Config:  config,
	}
}

func parseExporterWorkerConfigParams(values url.Values) *ExporterWorkerConfigParams {
	config := ExporterWorkerConfigParams{}

	exportConfigs := make(map[string]*SingleExportConfigParams)

	for key, values := range values {
		parts := strings.Split(key, "_")
		if len(parts) != 2 {
			continue
		}
		key := parts[1]
		exportConfig, e := exportConfigs[key]
		if !e {
			exportConfig = &SingleExportConfigParams{}
			exportConfigs[key] = exportConfig
		}
		switch parts[0] {
		case "storage":
			exportConfig.Storage = values[0]
		case "item":
			exportConfig.Item = values[0]
		case "slot":
			exportConfig.Slot = values[0]
		case "amount":
			exportConfig.Amount = values[0]
		}
	}

	keys := common.MapKeys(exportConfigs)
	sort.Strings(keys)

	for _, key := range keys {
		exportConfig := exportConfigs[key]
		config.Exports = append(config.Exports, *exportConfig)
	}
	return &config
}

func parseImporterWorkerConfigParams(values url.Values) *ImporterWorkerConfigParams {
	config := ImporterWorkerConfigParams{}

	importConfigs := make(map[string]*SingleImportConfigParams)

	for key, values := range values {
		parts := strings.Split(key, "_")
		if len(parts) != 2 {
			continue
		}
		key := parts[1]
		importConfig, e := importConfigs[key]
		if !e {
			importConfig = &SingleImportConfigParams{}
			importConfigs[key] = importConfig
		}
		switch parts[0] {
		case "storage":
			importConfig.Storage = values[0]
		case "slot":
			importConfig.Slot = values[0]
		}
	}

	for _, importConfig := range importConfigs {
		config.Imports = append(config.Imports, *importConfig)
	}
	return &config
}

func parseProcessingCrafterWorkerConfigParams(values url.Values) *dao.ProcessingCrafterWorkerConfig {
	config := dao.ProcessingCrafterWorkerConfig{}

	config.CraftType = values.Get("craftType")
	config.InputStorage = values.Get("inputStorage")
	config.ReagentMode = values.Get("reagentMode")

	return &config
}

func NewWorkerParams(worker *dao.Worker) *WorkerParams {
	config := WorkerConfigParams{}

	switch worker.Type {
	case dao.WORKER_TYPE_EXPORTER:
		exporterConfig := &ExporterWorkerConfigParams{}
		for _, exportConfig := range worker.Config.Exporter.Exports {
			exporterConfig.Exports = append(exporterConfig.Exports, SingleExportConfigParams{
				Storage: exportConfig.Storage,
				Item:    exportConfig.Item,
				Slot:    strconv.Itoa(exportConfig.Slot),
				Amount:  strconv.Itoa(exportConfig.Amount),
			})
		}
		config.Exporter = exporterConfig

	case dao.WORKER_TYPE_IMPORTER:
		importerConfig := &ImporterWorkerConfigParams{}
		for _, importConfig := range worker.Config.Importer.Imports {
			importerConfig.Imports = append(importerConfig.Imports, SingleImportConfigParams{
				Storage: importConfig.Storage,
				Slot:    strconv.Itoa(importConfig.Slot),
			})
		}
		config.Importer = importerConfig
	case dao.WORKER_TYPE_PROCESSING_CRAFTER:
		config.ProcessingCrafter = worker.Config.ProcessingCrafter
	}

	return &WorkerParams{
		Key:     worker.Key,
		Type:    worker.Type,
		Enabled: worker.Enabled,
		Config:  config,
	}
}

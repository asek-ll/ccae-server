package dao

import (
	"database/sql"
	"encoding/json"
	"errors"
)

var WORKER_TYPE_SHAPED_CRAFTER = "shaped_crafter"
var WORKER_TYPE_PROCESSING_CRAFTER = "processing_crafter"
var WORKER_TYPE_IMPORTER = "importer"
var WORKER_TYPE_EXPORTER = "exporter"

var WORKER_TYPES = []string{
	WORKER_TYPE_SHAPED_CRAFTER,
	WORKER_TYPE_PROCESSING_CRAFTER,
	WORKER_TYPE_IMPORTER,
	WORKER_TYPE_EXPORTER,
}

type SingleImportConfig struct {
	Storage string `json:"storage"`
	Slot    int    `json:"slot"`
}

type ImporterWorkerConfig struct {
	Imports []SingleImportConfig `json:"imports"`
}

type SingleExportConfig struct {
	Storage string `json:"storage"`
	Item    string `json:"item"`
	Slot    int    `json:"slot"`
	Amount  int    `json:"amount"`
}

type ExporterWorkerConfig struct {
	Exports []SingleExportConfig `json:"exports"`
}
type ShapedCrafterWorkerConfig struct {
	InputStorage  string `json:"inputStorage"`
	OutputStorage string `json:"outputStorage"`
}
type ProcessingCrafterWorkerConfig struct {
	CraftType    string `json:"craftType"`
	InputStorage string `json:"inputStorage"`
	InputTank    string `json:"inputTank"`
	ReagentMode  string `json:"reagentMode"`
}

type WorkerConfig struct {
	ShapedCrafter     *ShapedCrafterWorkerConfig     `json:"shapedCrafter"`
	ProcessingCrafter *ProcessingCrafterWorkerConfig `json:"processingCrafter"`
	Importer          *ImporterWorkerConfig          `json:"importer"`
	Exporter          *ExporterWorkerConfig          `json:"exporter"`
}

type Worker struct {
	Key     string
	Type    string
	Enabled bool
	Config  WorkerConfig
}

type WorkersDao struct {
	db *sql.DB
}

func NewWorkersDao(db *sql.DB) (*WorkersDao, error) {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS worker (
		key string NOT NULL PRIMARY KEY,
		type string NOT NULL,
		enabled bool NOT NULL,
		config string NOT NULL
	);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &WorkersDao{
		db: db,
	}, nil
}

func (w *WorkersDao) GetWorkers() ([]Worker, error) {
	rows, err := w.db.Query("SELECT key, type, enabled, config FROM worker")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readWorkers(rows)
}

func parseConfig(config string) (*WorkerConfig, error) {
	if config == "" {
		return nil, nil
	}
	var result WorkerConfig
	err := json.Unmarshal([]byte(config), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func readWorkers(rows *sql.Rows) ([]Worker, error) {
	var workers []Worker
	for rows.Next() {
		var w Worker
		var config string
		err := rows.Scan(&w.Key, &w.Type, &w.Enabled, &config)
		if err != nil {
			return nil, err
		}
		parsedConfig, err := parseConfig(config)
		if err != nil {
			return nil, err
		}
		if parsedConfig != nil {
			w.Config = *parsedConfig
		}
		workers = append(workers, w)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return workers, nil
}

func (w *WorkersDao) GetWorker(key string) (*Worker, error) {
	rows, err := w.db.Query("SELECT key, type, enabled, config FROM worker WHERE key = ?", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workers, err := readWorkers(rows)
	if err != nil {
		return nil, err
	}
	if len(workers) == 0 {
		return nil, errors.New("Worker not found")
	}
	return &workers[0], nil
}

func (w *WorkersDao) CreateWorker(worker *Worker) error {
	config, err := json.Marshal(worker.Config)
	if err != nil {
		return err
	}
	_, err = w.db.Exec("INSERT INTO worker (key, type, enabled, config) VALUES (?, ?, ?, ?)", worker.Key, worker.Type, worker.Enabled, string(config))
	return err
}

func (w *WorkersDao) UpdateWorker(key string, worker *Worker) error {
	config, err := json.Marshal(worker.Config)
	if err != nil {
		return err
	}
	_, err = w.db.Exec("UPDATE worker SET key = ?, type = ?, enabled = ?, config = ? WHERE key = ?", worker.Key, worker.Type, worker.Enabled, string(config), key)
	return err
}

func (w *WorkersDao) DeleteWorker(key string) error {
	_, err := w.db.Exec("DELETE FROM worker WHERE key = ?", key)
	return err
}

package config

type AppConfig struct {
	Storage   StorageConfig   `json:"storage"`
	Crafters  CraftersConfig  `json:"crafters"`
	Importers ImportersConfig `json:"importers"`
}

type StorageConfig struct {
}

type CraftersConfig struct {
	ProcessCrafters []ProcessCrafterConfig `json:"processCrafters"`
}

type ImportersConfig struct {
	StorageImporters []StorageImporterConfig `json:"storage"`
}

type ProcessCrafterConfig struct {
	CraftType      string `json:"craftType"`
	InputInventory string `json:"inputInventory"`
	InputTank      string `json:"inputTank"`
	ReagentMode    string `json:"reagentMode"`
	Enabled        bool   `json:"enabled"`
}

type StorageImporterConfig struct {
	Storage string   `json:"storage"`
	Items   []string `json:"items"`
}

// type StorageExporterConfig struct {
// 	Storage string   `json:"storage"`
// 	Items   []string `json:"items"`
// }

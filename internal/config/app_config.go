package config

type AppConfig struct {
	Storage   StorageConfig   `json:"storage"`
	Crafters  CraftersConfig  `json:"crafters"`
	Importers ImportersConfig `json:"importers"`
}

type StorageConfig struct {
	ColdStoragePrefix          string `json:"coldStoragePrefix"`
	WarmStoragePrefix          string `json:"warmStoragePrefix"`
	SingleFluidContainerPrefix string `json:"fluidContainerPrefix"`

	InputStorages []string `json:"inputStorages"`
}

type CraftersConfig struct {
	ProcessCrafters []ProcessCrafterConfig `json:"processCrafters"`
}

type ImportersConfig struct {
	StorageImporters []StorageImporterConfig `json:"storage"`
	FluidImporters   []FluidImporterConfig   `json:"fluid"`
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

type FluidImporterConfig struct {
	Tank   string   `json:"tank"`
	Fluids []string `json:"fluids"`
}

// type StorageExporterConfig struct {
// 	Storage string   `json:"storage"`
// 	Items   []string `json:"items"`
// }

// InputStorages              []string
// ColdStoragePrefix          string
// WarmStoragePrefix          string
// SingleFluidContainerPrefix string
// TransactionStorage         string
// TransactionTank            string

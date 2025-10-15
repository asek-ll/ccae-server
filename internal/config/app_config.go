package config

type AppConfig struct {
	Storage   StorageConfig   `json:"storage"`
	Crafters  CraftersConfig  `json:"crafters"`
	Importers ImportersConfig `json:"importers"`

	WebServer    WebServerConfig    `json:"webServer"`
	ClientServer ClientServerConfig `json:"clientServer"`
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

type ProcessCrafterConfig struct {
	WorkerKey      string `json:"workerKey"`
	CraftType      string `json:"craftType"`
	InputInventory string `json:"inputInventory"`
	InputTank      string `json:"inputTank"`
	ReagentMode    string `json:"reagentMode"`
	Enabled        bool   `json:"enabled"`
	WaitResults    bool   `json:"waitResults"`
	CraftCondition string `json:"craftCondition"`

	ResultItems          []string `json:"resultItems"`
	ResultInventory      string   `json:"resultInventory"`
	ResultInventorySlots []int    `json:"resultInventorySlots"`

	ResultFluids []string `json:"resultFluids"`
	ResultTank   string   `json:"resultTank"`
}

type ImportersConfig struct {
	StorageImporters []StorageImporterConfig `json:"storage"`
	FluidImporters   []FluidImporterConfig   `json:"fluid"`
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

type AuthConfig struct {
	OAuthClient   string   `json:"oauthClient"`
	OAuthSecret   string   `json:"oauthSecret"`
	TokenSecret   string   `json:"tokenSecret"`
	Admins        []string `json:"admins"`
	AdminPassword string   `json:"adminPassword"`
}

type ClientServerConfig struct {
	Url        string `json:"url"`
	ListenAddr string `json:"addr"`
}

type WebServerConfig struct {
	Url        string     `json:"url"`
	ListenAddr string     `json:"addr"`
	Auth       AuthConfig `json:"auth"`
}

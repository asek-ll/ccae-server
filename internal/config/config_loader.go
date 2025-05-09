package config

import (
	"encoding/json"
	"os"
)

type ConfigLoader struct {
	Config AppConfig
}

func NewConfigLoader(configFile string) (*ConfigLoader, error) {
	jsonFile, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigLoader{}, nil
		}
		return nil, err
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	var config AppConfig
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &ConfigLoader{
		Config: config,
	}, nil
}

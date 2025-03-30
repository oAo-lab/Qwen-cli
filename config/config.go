package config

import (
	"encoding/json"
	"os"
)

type ModelConfig struct {
	Name string `json:"name"`
}

type Config struct {
	APIURL string                 `json:"api_url"`
	APIKey string                 `json:"api_key"`
	Models map[string]ModelConfig `json:"models"`
	Roles  map[string]string      `json:"roles"`
}

func LoadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

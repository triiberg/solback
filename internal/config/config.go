package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBDSN        string `json:"db_dsn"`
	OpenAIAPIKey string `json:"openai_api_key"`
}

func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("config path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("db_dsn is required")
	}
	if cfg.OpenAIAPIKey == "" {
		return Config{}, fmt.Errorf("openai_api_key is required")
	}

	return cfg, nil
}

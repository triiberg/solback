package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type SourceConfig struct {
	Source SourceEntry `json:"source"`
}

type SourceEntry struct {
	URL     string `json:"url"`
	Comment string `json:"comment"`
}

func LoadSourceConfig(path string) (SourceConfig, error) {
	if path == "" {
		return SourceConfig{}, fmt.Errorf("source config path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return SourceConfig{}, fmt.Errorf("read source config: %w", err)
	}

	var cfg SourceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return SourceConfig{}, fmt.Errorf("parse source config: %w", err)
	}

	if cfg.Source.URL == "" {
		return SourceConfig{}, fmt.Errorf("source.url is required")
	}
	if cfg.Source.Comment == "" {
		return SourceConfig{}, fmt.Errorf("source.comment is required")
	}

	return cfg, nil
}

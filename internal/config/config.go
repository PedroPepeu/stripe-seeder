package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".stripe-seeder.json"

type Config struct {
	APIKey      string `json:"api_key"`
	ProjectName string `json:"project_name"`
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return configFileName
	}
	return filepath.Join(home, configFileName)
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, nil
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

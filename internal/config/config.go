// Package config provides application configuration loading and validation.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration.
// It is structured to support multiple service configurations.
type Config struct {
	YNAB YNABConfig `json:"ynab"`
	// Future configurations can be added here:
	// Sparkasse SparkasseConfig `json:"sparkasse"`
}

// YNABConfig holds configuration for the YNAB API client.
type YNABConfig struct {
	// APIKey is the personal access token for YNAB API authentication.
	APIKey string `json:"api_key"`
	// BudgetID is the default budget to use for API operations.
	BudgetID string `json:"budget_id"`
}

// LoadFromFile reads and parses a JSON configuration file from the given path.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration contains all required fields.
func (c *Config) Validate() error {
	if err := c.YNAB.Validate(); err != nil {
		return fmt.Errorf("ynab config: %w", err)
	}
	return nil
}

// Validate checks that the YNAB configuration contains all required fields.
func (y *YNABConfig) Validate() error {
	if y.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if y.BudgetID == "" {
		return fmt.Errorf("budget_id is required")
	}
	return nil
}

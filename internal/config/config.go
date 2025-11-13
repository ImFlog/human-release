package config

import (
	"fmt"
	"os"

	"github.com/ImFlog/human-release/pkg/models"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration from TOML file and environment variables
func LoadConfig(configPath string) (*models.Config, error) {
	v := viper.New()

	// Set environment variable prefix for auto-binding
	v.SetEnvPrefix("HUMAN_RELEASE")
	v.AutomaticEnv()

	// Bind specific environment variables
	v.BindEnv("llm.api_key", "HUMAN_RELEASE_LLM_API_KEY")
	v.BindEnv("llm.base_url", "HUMAN_RELEASE_LLM_BASE_URL")

	// Load from config file if provided
	if configPath != "" {
		// Check if file exists
		if _, err := os.Stat(configPath); err == nil {
			v.SetConfigFile(configPath)
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Unmarshal into config struct
	cfg := &models.Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return cfg, nil
}

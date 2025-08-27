package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the user's local configuration
type Config struct {
	APIKey                string   `yaml:"api_key" json:"api_key"`
	DefaultPort           int      `yaml:"default_port" json:"default_port"`
	DefaultConcurrency    int      `yaml:"default_concurrency" json:"default_concurrency"`
	DefaultMaxPages       int      `yaml:"default_max_pages" json:"default_max_pages"`
	DefaultMaxDepth       int      `yaml:"default_max_depth" json:"default_max_depth"`
	DefaultIgnorePatterns []string `yaml:"default_ignore_patterns" json:"default_ignore_patterns"`
}

// GetEffectiveBaseURL returns the base URL to use, checking environment variable override first
func (c *Config) GetEffectiveBaseURL() string {
	// Check for environment variable override first (for development only)
	if envURL := os.Getenv("SEO_BASE_URL"); envURL != "" {
		return envURL
	}

	// Always use production URL for users
	return "https://seofor.dev"
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".seo")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.yml"), nil
}

// LoadConfig loads configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return GetDefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Fill in defaults for missing fields
	fillDefaults(&config)
	
	return &config, nil
}

// SaveConfig saves configuration to the config file
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// LoadOrCreateConfig loads existing config or creates a new one with defaults
func LoadOrCreateConfig() (*Config, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// If this is a fresh config (no API key), we might want to save defaults
	if config.APIKey == "" {
		fillDefaults(config)
	}

	return config, nil
}

// GetDefaultConfig returns a config with default values
func GetDefaultConfig() *Config {
	config := &Config{
		APIKey:             "",
		DefaultPort:        3000,
		DefaultConcurrency: 4,
		DefaultMaxPages:    0,
		DefaultMaxDepth:    0,
		DefaultIgnorePatterns: []string{
			"/api",
			"/admin",
		},
	}
	
	return config
}

// fillDefaults fills in any missing configuration with defaults
func fillDefaults(config *Config) {
	if config.DefaultPort == 0 {
		config.DefaultPort = 3000
	}
	if config.DefaultConcurrency == 0 {
		config.DefaultConcurrency = 4
	}
	if config.DefaultIgnorePatterns == nil {
		config.DefaultIgnorePatterns = []string{"/api", "/admin"}
	}
}

// ValidateAPIKeyWithMessage checks if API key is configured and shows setup message if not
func ValidateAPIKeyWithMessage() (*Config, error) {
	config, err := LoadOrCreateConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf(`API key required for this operation.

Premium API key setup:
  1. Subscribe to a plan: https://seofor.dev/payments/pricing
  2. Get your API key: https://seofor.dev/dashboard  
  3. Set your API key: seo config set-api-key <your-key>

Run 'seo pro setup' for detailed setup instructions.`)
	}

	return config, nil
}
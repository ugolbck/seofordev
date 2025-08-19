package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Constants
const (
	BaseURL = "https://seofor.dev"
)

// Config represents the user's local configuration
type Config struct {
	APIKey                string   `yaml:"api_key" json:"api_key"`
	DefaultPort           int      `yaml:"default_port" json:"default_port"`
	DefaultConcurrency    int      `yaml:"default_concurrency" json:"default_concurrency"`
	DefaultMaxPages       int      `yaml:"default_max_pages" json:"default_max_pages"`
	DefaultMaxDepth       int      `yaml:"default_max_depth" json:"default_max_depth"`
	DefaultIgnorePatterns []string `yaml:"default_ignore_patterns" json:"default_ignore_patterns"`
	FollowRobotsTxt       bool     `yaml:"follow_robots_txt" json:"follow_robots_txt"`
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
		return GetDefaultConfig(), err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return GetDefaultConfig(), err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return GetDefaultConfig(), err
	}

	// Ensure we have default values for missing fields
	defaultConfig := GetDefaultConfig()
	if config.DefaultPort == 0 {
		config.DefaultPort = defaultConfig.DefaultPort
	}
	if config.DefaultConcurrency == 0 {
		config.DefaultConcurrency = defaultConfig.DefaultConcurrency
	}
	if config.DefaultMaxPages == 0 {
		config.DefaultMaxPages = defaultConfig.DefaultMaxPages
	}
	if config.DefaultMaxDepth == 0 {
		config.DefaultMaxDepth = defaultConfig.DefaultMaxDepth
	}
	if config.DefaultIgnorePatterns == nil {
		config.DefaultIgnorePatterns = defaultConfig.DefaultIgnorePatterns
	}

	return &config, nil
}

// GetEffectiveBaseURL returns the base URL to use, checking environment variable override first
func (c *Config) GetEffectiveBaseURL() string {
	// Check for environment variable override first (for development only)
	if envURL := os.Getenv("SEO_BASE_URL"); envURL != "" {
		return envURL
	}

	// Always use production URL for users
	return BaseURL
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

	// Write with user-only permissions for security
	return os.WriteFile(configPath, data, 0600)
}

// GetDefaultConfig returns default configuration values
func GetDefaultConfig() *Config {
	return &Config{
		APIKey:                "",
		DefaultPort:           3000,
		DefaultConcurrency:    4,
		DefaultMaxPages:       0,                          // unlimited
		DefaultMaxDepth:       0,                          // unlimited
		DefaultIgnorePatterns: []string{"/api", "/admin"}, // Default ignore patterns
	}
}

// MaskAPIKey returns a partially masked version of the API key for display
func MaskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}

	if len(apiKey) <= 10 {
		// For short keys, show just first and last character
		if len(apiKey) <= 2 {
			return strings.Repeat("*", len(apiKey))
		}
		return string(apiKey[0]) + strings.Repeat("*", len(apiKey)-2) + string(apiKey[len(apiKey)-1])
	}

	// For longer keys (like OpenAI keys), show first 3 and last 6 characters
	return apiKey[:3] + "..." + apiKey[len(apiKey)-6:]
}

// APIValidationResponse represents the response from the API validation endpoint
type APIValidationResponse struct {
	Valid bool `json:"valid"`
	User  struct {
		ID          int     `json:"id"`
		Email       string  `json:"email"`
		Username    *string `json:"username"` // nullable
		Credits     int     `json:"credits"`
		HasPaidPlan bool    `json:"has_paid_plan"`
		IsVerified  bool    `json:"is_verified"`
		CreatedAt   *string `json:"created_at"` // nullable
	} `json:"user"`
}

// ValidateAPIKeyWithServer validates the API key with the server
func ValidateAPIKeyWithServer(apiKey string, baseURL string) (*APIValidationResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request to validation endpoint
	validationURL := fmt.Sprintf("%s/api/auth/validate/", baseURL)
	req, err := http.NewRequest("GET", validationURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("User-Agent", "SEO-CLI/2.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API key validation failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var validationResp APIValidationResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	if !validationResp.Valid {
		return nil, fmt.Errorf("API key is not valid")
	}

	return &validationResp, nil
}

// ValidateAPIKey performs basic validation on the API key
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if len(apiKey) < 10 {
		return fmt.Errorf("API key seems too short")
	}

	// Add more validation rules as needed
	return nil
}

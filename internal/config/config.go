package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CloudConfig struct {
	Provider       string            `json:"provider"`
	Region         string            `json:"region"`
	Endpoint       string            `json:"endpoint"`
	AccessKey      string            `json:"access_key"`
	SecretKey      string            `json:"secret_key"`
	Bucket         string            `json:"bucket"`
	AccountID      string            `json:"account_id,omitempty"`
	CustomDomain   string            `json:"custom_domain,omitempty"`
	PathPrefix     string            `json:"path_prefix,omitempty"`
	ProviderOpts   map[string]string `json:"provider_opts,omitempty"`
	Concurrency    int               `json:"concurrency"`
	SkipVerify     bool              `json:"skip_verify"`
	CACertPath     string            `json:"ca_cert_path,omitempty"`
	ConflictPolicy string            `json:"conflict_policy"`
	CacheDir       string            `json:"cache_dir,omitempty"`
}

type Config struct {
	TranslatePrompt   string                 `json:"translate_prompt"`
	OpenAIEndpointURL string                 `json:"endpoint"`
	OpenAIAPIKey      string                 `json:"api_key"`
	ModelName         string                 `json:"model"`
	Temperature       float64                `json:"temperature"`
	TopP              float64                `json:"top_p"`
	CloudStorages     map[string]CloudConfig `json:"cloud_storages,omitempty"`
	DefaultStorage    string                 `json:"default_storage,omitempty"`
}

var DefaultCloudConfig = CloudConfig{
	Provider:       "",
	Region:         "auto",
	Endpoint:       "",
	AccessKey:      "",
	SecretKey:      "",
	Bucket:         "",
	Concurrency:    5,
	SkipVerify:     false,
	ConflictPolicy: "rename",
}

var DefaultConfig = Config{
	TranslatePrompt:   "Translate the markdown to {TARGET_LANG} as a native speaker - preserve code/YAML/links/cli commands (e.g. `kubectl apply` or `pip install langchain`) and tech terms (CRDs, Helm charts, RAG). Output ONLY fluently localized text with natural technical phrasing that doesn't read machine-generated.",
	OpenAIEndpointURL: "https://api.openai.com/v1",
	OpenAIAPIKey:      "",
	ModelName:         "gpt-3.5-turbo",
	Temperature:       0.0,
	TopP:              1.0,
	CloudStorages:     make(map[string]CloudConfig),
}

func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "mdctl", "config.json")
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()
	if configPath == "" {
		return &DefaultConfig, nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := SaveConfig(&DefaultConfig); err != nil {
			return &DefaultConfig, fmt.Errorf("failed to create default config: %v", err)
		}
		return &DefaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &DefaultConfig, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		os.Remove(configPath)
		if err := SaveConfig(&DefaultConfig); err != nil {
			return &DefaultConfig, fmt.Errorf("failed to create new config after invalid file: %v", err)
		}
		return &DefaultConfig, fmt.Errorf("invalid config file (recreated with defaults): %v", err)
	}

	if config.TranslatePrompt == "" {
		config.TranslatePrompt = DefaultConfig.TranslatePrompt
	}
	if config.OpenAIEndpointURL == "" {
		config.OpenAIEndpointURL = DefaultConfig.OpenAIEndpointURL
	}
	if config.ModelName == "" {
		config.ModelName = DefaultConfig.ModelName
	}

	// Ensure CloudStorages is non-nil
	if config.CloudStorages == nil {
		config.CloudStorages = make(map[string]CloudConfig)
	}

	// Check if default storage exists
	if config.DefaultStorage != "" {
		if _, exists := config.CloudStorages[config.DefaultStorage]; !exists {
			// If specified default storage doesn't exist, use the first available one
			if len(config.CloudStorages) > 0 {
				for name := range config.CloudStorages {
					config.DefaultStorage = name
					break
				}
			} else {
				config.DefaultStorage = ""
			}
		}
	} else if len(config.CloudStorages) > 0 {
		// If no default storage is set but there are storage configurations, set the first one as default
		for name := range config.CloudStorages {
			config.DefaultStorage = name
			break
		}
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath := GetConfigPath()
	if configPath == "" {
		return fmt.Errorf("failed to get config path")
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// ApplyCloudConfig applies platform-specific settings to the cloud configuration
func (c *Config) ApplyCloudConfig() {
	// Ensure CloudStorages is non-nil
	if c.CloudStorages == nil {
		c.CloudStorages = make(map[string]CloudConfig)
	}

	// Check if default storage exists
	if c.DefaultStorage != "" {
		if _, exists := c.CloudStorages[c.DefaultStorage]; !exists {
			// If specified default storage doesn't exist, use the first available one
			for name := range c.CloudStorages {
				c.DefaultStorage = name
				break
			}
		}
	}

	// If no default storage is set but there are storage configurations, set the first one as default
	if c.DefaultStorage == "" && len(c.CloudStorages) > 0 {
		for name := range c.CloudStorages {
			c.DefaultStorage = name
			break
		}
	}
}

// GetActiveCloudConfig returns the current active cloud storage configuration
// The storageName parameter can specify which configuration to use, if empty the default configuration is used
func (c *Config) GetActiveCloudConfig(storageName string) CloudConfig {
	// If a storage name is specified, try to get that configuration
	if storageName != "" {
		if storage, exists := c.CloudStorages[storageName]; exists {
			return storage
		}
	}

	// If there's a default configuration, use that
	if c.DefaultStorage != "" {
		if storage, exists := c.CloudStorages[c.DefaultStorage]; exists {
			return storage
		}
	}

	// If any configuration is available, return the first one found
	if len(c.CloudStorages) > 0 {
		for _, storage := range c.CloudStorages {
			return storage
		}
	}

	// Return default empty configuration
	return DefaultCloudConfig
}

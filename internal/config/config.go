package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	TranslatePrompt   string  `json:"translate_prompt"`
	OpenAIEndpointURL string  `json:"endpoint"`
	OpenAIAPIKey      string  `json:"api_key"`
	ModelName         string  `json:"model"`
	Temperature       float64 `json:"temperature"`
	TopP              float64 `json:"top_p"`
}

var DefaultConfig = Config{
	TranslatePrompt:   "Please translate the following markdown content to {TARGET_LANG}. Keep the original markdown format and front matter unchanged. Do not add any additional markdown code blocks or backticks. Translate the content directly:",
	OpenAIEndpointURL: "https://api.openai.com/v1",
	OpenAIAPIKey:      "",
	ModelName:         "gpt-3.5-turbo",
	Temperature:       0.0,
	TopP:              1.0,
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

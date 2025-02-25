package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samzong/mdctl/internal/config"
	"github.com/spf13/cobra"
)

var (
	configKey   string
	configValue string
	storageName string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and modify configuration settings`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		// Create a temporary struct to control JSON output
		type ConfigDisplay struct {
			TranslatePrompt   string                        `json:"translate_prompt"`
			OpenAIEndpointURL string                        `json:"endpoint"`
			OpenAIAPIKey      string                        `json:"api_key"`
			ModelName         string                        `json:"model"`
			Temperature       float64                       `json:"temperature"`
			TopP              float64                       `json:"top_p"`
			CloudStorages     map[string]config.CloudConfig `json:"cloud_storages,omitempty"`
			DefaultStorage    string                        `json:"default_storage,omitempty"`
		}

		display := ConfigDisplay{
			TranslatePrompt:   cfg.TranslatePrompt,
			OpenAIEndpointURL: cfg.OpenAIEndpointURL,
			OpenAIAPIKey:      cfg.OpenAIAPIKey,
			ModelName:         cfg.ModelName,
			Temperature:       cfg.Temperature,
			TopP:              cfg.TopP,
			CloudStorages:     cfg.CloudStorages,
			DefaultStorage:    cfg.DefaultStorage,
		}

		data, err := json.MarshalIndent(display, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
	Example: `  mdctl config set --key api_key --value "your-api-key"
  mdctl config set --key model --value "gpt-4"
  mdctl config set --key temperature --value "0.8"
  
  # Cloud storage configuration
  mdctl config set --key cloud_storages.my-s3.provider --value "s3"
  mdctl config set --key cloud_storages.my-r2.provider --value "r2"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configKey == "" {
			return fmt.Errorf("key is required")
		}
		if configValue == "" {
			return fmt.Errorf("value is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		// Handle multi-cloud storage configurations with cloud_storages.<n>.<field>
		if strings.HasPrefix(strings.ToLower(configKey), "cloud_storages.") {
			parts := strings.SplitN(configKey, ".", 3)
			if len(parts) != 3 {
				return fmt.Errorf("invalid config key format: %s", configKey)
			}

			// Ensure CloudStorages map is initialized
			if cfg.CloudStorages == nil {
				cfg.CloudStorages = make(map[string]config.CloudConfig)
			}

			storageName := parts[1]
			field := parts[2]

			// Get or create storage configuration
			storage, exists := cfg.CloudStorages[storageName]
			if !exists {
				storage = config.DefaultCloudConfig
			}

			// Set field value
			switch strings.ToLower(field) {
			case "provider":
				storage.Provider = configValue
			case "region":
				storage.Region = configValue
			case "endpoint":
				storage.Endpoint = configValue
			case "access_key":
				storage.AccessKey = configValue
			case "secret_key":
				storage.SecretKey = configValue
			case "bucket":
				storage.Bucket = configValue
			case "account_id":
				storage.AccountID = configValue
			case "custom_domain":
				storage.CustomDomain = configValue
			case "path_prefix":
				storage.PathPrefix = configValue
			case "concurrency":
				var concurrency int
				if _, err := fmt.Sscanf(configValue, "%d", &concurrency); err != nil {
					return fmt.Errorf("invalid concurrency value: %s", configValue)
				}
				storage.Concurrency = concurrency
			case "skip_verify":
				skipVerify := strings.ToLower(configValue) == "true"
				storage.SkipVerify = skipVerify
			case "ca_cert_path":
				storage.CACertPath = configValue
			case "conflict_policy":
				policy := strings.ToLower(configValue)
				if policy != "rename" && policy != "version" && policy != "overwrite" {
					return fmt.Errorf("invalid conflict policy: %s (must be rename, version, or overwrite)", configValue)
				}
				storage.ConflictPolicy = policy
			case "cache_dir":
				storage.CacheDir = configValue
			default:
				return fmt.Errorf("unknown cloud storage configuration key: %s", field)
			}

			// Save the updated storage configuration
			cfg.CloudStorages[storageName] = storage

			// If default storage is not set and there's only one storage, set it as default
			if cfg.DefaultStorage == "" && len(cfg.CloudStorages) == 1 {
				cfg.DefaultStorage = storageName
			}

		} else {
			// Handle existing config settings
			switch strings.ToLower(configKey) {
			case "translate_prompt":
				cfg.TranslatePrompt = configValue
			case "endpoint":
				cfg.OpenAIEndpointURL = configValue
			case "api_key":
				cfg.OpenAIAPIKey = configValue
			case "model":
				cfg.ModelName = configValue
			case "temperature":
				var temp float64
				if _, err := fmt.Sscanf(configValue, "%f", &temp); err != nil {
					return fmt.Errorf("invalid temperature value: %s", configValue)
				}
				cfg.Temperature = temp
			case "top_p":
				var topP float64
				if _, err := fmt.Sscanf(configValue, "%f", &topP); err != nil {
					return fmt.Errorf("invalid top_p value: %s", configValue)
				}
				cfg.TopP = topP
			default:
				return fmt.Errorf("unknown configuration key: %s", configKey)
			}
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}

		fmt.Printf("Successfully set %s to %s\n", configKey, configValue)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a configuration value",
	Example: `  mdctl config get --key api_key
  mdctl config get --key model
  mdctl config get --key cloud_storages.my-r2.provider`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configKey == "" {
			return fmt.Errorf("key is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		// Handle cloud storage configurations
		if strings.HasPrefix(strings.ToLower(configKey), "cloud_storages.") {
			parts := strings.SplitN(configKey, ".", 3)
			if len(parts) != 3 {
				return fmt.Errorf("invalid cloud storages key format: %s, expected cloud_storages.<name>.<field>", configKey)
			}

			storageName := parts[1]
			field := parts[2]

			// Check if specified storage exists
			if _, exists := cfg.CloudStorages[storageName]; !exists {
				return fmt.Errorf("storage '%s' does not exist", storageName)
			}

			var value interface{}
			switch strings.ToLower(field) {
			case "provider":
				value = cfg.CloudStorages[storageName].Provider
			case "region":
				value = cfg.CloudStorages[storageName].Region
			case "endpoint":
				value = cfg.CloudStorages[storageName].Endpoint
			case "access_key":
				value = cfg.CloudStorages[storageName].AccessKey
			case "secret_key":
				value = cfg.CloudStorages[storageName].SecretKey
			case "bucket":
				value = cfg.CloudStorages[storageName].Bucket
			case "account_id":
				value = cfg.CloudStorages[storageName].AccountID
			case "custom_domain":
				value = cfg.CloudStorages[storageName].CustomDomain
			case "path_prefix":
				value = cfg.CloudStorages[storageName].PathPrefix
			case "concurrency":
				value = cfg.CloudStorages[storageName].Concurrency
			case "skip_verify":
				value = cfg.CloudStorages[storageName].SkipVerify
			case "ca_cert_path":
				value = cfg.CloudStorages[storageName].CACertPath
			case "conflict_policy":
				value = cfg.CloudStorages[storageName].ConflictPolicy
			case "cache_dir":
				value = cfg.CloudStorages[storageName].CacheDir
			default:
				return fmt.Errorf("unknown cloud storage configuration key: %s", field)
			}

			fmt.Printf("%v\n", value)
			return nil
		}

		// Handle existing config settings
		var value interface{}
		switch strings.ToLower(configKey) {
		case "translate_prompt":
			value = cfg.TranslatePrompt
		case "endpoint":
			value = cfg.OpenAIEndpointURL
		case "api_key":
			value = cfg.OpenAIAPIKey
		case "model":
			value = cfg.ModelName
		case "temperature":
			value = cfg.Temperature
		case "top_p":
			value = cfg.TopP
		default:
			return fmt.Errorf("unknown configuration key: %s", configKey)
		}

		fmt.Printf("%v\n", value)
		return nil
	},
}

var configSetDefaultStorageCmd = &cobra.Command{
	Use:     "set-default-storage",
	Short:   "Set the default cloud storage configuration",
	Example: `  mdctl config set-default-storage --name my-r2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if storageName == "" {
			return fmt.Errorf("storage name is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		// Check if specified storage exists
		if _, exists := cfg.CloudStorages[storageName]; !exists {
			return fmt.Errorf("storage '%s' does not exist", storageName)
		}

		cfg.DefaultStorage = storageName

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}

		fmt.Printf("Default storage set to: %s\n", storageName)
		return nil
	},
}

var configListStoragesCmd = &cobra.Command{
	Use:   "list-storages",
	Short: "List all cloud storage configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		fmt.Println("Cloud Storage Configurations:")
		fmt.Println("-----------------------------")

		// List multi-cloud storage configurations
		if len(cfg.CloudStorages) > 0 {
			for name, storage := range cfg.CloudStorages {
				isDefault := name == cfg.DefaultStorage
				defaultMark := ""
				if isDefault {
					defaultMark = " (DEFAULT)"
				}
				fmt.Printf("Storage: %s%s\n", name, defaultMark)
				data, err := json.MarshalIndent(storage, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal config: %v", err)
				}
				fmt.Println(string(data))
				fmt.Println()
			}
		} else {
			fmt.Println("No cloud storage configurations found.")
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetDefaultStorageCmd)
	configCmd.AddCommand(configListStoragesCmd)

	configSetCmd.Flags().StringVarP(&configKey, "key", "k", "", "Configuration key to set")
	configSetCmd.Flags().StringVarP(&configValue, "value", "v", "", "Value to set")
	configSetCmd.MarkFlagRequired("key")
	configSetCmd.MarkFlagRequired("value")

	configGetCmd.Flags().StringVarP(&configKey, "key", "k", "", "Configuration key to get")
	configGetCmd.MarkFlagRequired("key")

	configSetDefaultStorageCmd.Flags().StringVarP(&storageName, "name", "n", "", "Storage name to set as default")
	configSetDefaultStorageCmd.MarkFlagRequired("name")
}

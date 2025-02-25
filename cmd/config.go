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

		data, err := json.MarshalIndent(cfg, "", "  ")
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
  mdctl config set --key temperature --value "0.8"`,
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
  mdctl config get --key model`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configKey == "" {
			return fmt.Errorf("key is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

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

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)

	configSetCmd.Flags().StringVarP(&configKey, "key", "k", "", "Configuration key to set")
	configSetCmd.Flags().StringVarP(&configValue, "value", "v", "", "Value to set")
	configSetCmd.MarkFlagRequired("key")
	configSetCmd.MarkFlagRequired("value")

	configGetCmd.Flags().StringVarP(&configKey, "key", "k", "", "Configuration key to get")
	configGetCmd.MarkFlagRequired("key")
}

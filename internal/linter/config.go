package linter

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ConfigFile represents a markdownlint configuration file
type ConfigFile struct {
	// Default configuration
	Default bool `json:"default,omitempty"`

	// Extends other configuration files
	Extends string `json:"extends,omitempty"`

	// Rule-specific configuration
	MD001 *RuleConfig `json:"MD001,omitempty"`
	MD003 *RuleConfig `json:"MD003,omitempty"`
	MD009 *RuleConfig `json:"MD009,omitempty"`
	MD010 *RuleConfig `json:"MD010,omitempty"`
	MD012 *RuleConfig `json:"MD012,omitempty"`
	MD013 *RuleConfig `json:"MD013,omitempty"`
	MD018 *RuleConfig `json:"MD018,omitempty"`
	MD019 *RuleConfig `json:"MD019,omitempty"`
	MD023 *RuleConfig `json:"MD023,omitempty"`
	MD032 *RuleConfig `json:"MD032,omitempty"`
	MD047 *RuleConfig `json:"MD047,omitempty"`
}

// RuleConfig represents configuration for a specific rule
type RuleConfig struct {
	// Whether the rule is enabled
	Enabled *bool `json:"enabled,omitempty"`

	// Rule-specific options
	Options map[string]interface{} `json:"options,omitempty"`
}

// LoadConfigFile loads configuration from a file
func LoadConfigFile(filename string) (*ConfigFile, error) {
	// Try to find config file if not specified
	if filename == "" {
		filename = findConfigFile()
	}

	if filename == "" {
		return &ConfigFile{Default: true}, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ApplyToRuleSet applies the configuration to a rule set
func (c *ConfigFile) ApplyToRuleSet(rs *RuleSet) {
	ruleConfigs := map[string]*RuleConfig{
		"MD001": c.MD001,
		"MD003": c.MD003,
		"MD009": c.MD009,
		"MD010": c.MD010,
		"MD012": c.MD012,
		"MD013": c.MD013,
		"MD018": c.MD018,
		"MD019": c.MD019,
		"MD023": c.MD023,
		"MD032": c.MD032,
		"MD047": c.MD047,
	}

	for ruleID, ruleConfig := range ruleConfigs {
		if ruleConfig != nil && ruleConfig.Enabled != nil {
			if rule, exists := rs.rules[ruleID]; exists {
				rule.SetEnabled(*ruleConfig.Enabled)
			}
		}
	}
}

// findConfigFile looks for common markdownlint config files
func findConfigFile() string {
	configFiles := []string{
		".markdownlint.json",
		".markdownlint.jsonc",
		".markdownlintrc",
		".markdownlintrc.json",
		".markdownlintrc.jsonc",
	}

	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	// Also check in home directory
	if home, err := os.UserHomeDir(); err == nil {
		for _, filename := range configFiles {
			fullPath := filepath.Join(home, filename)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}

	return ""
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(filename string) error {
	config := ConfigFile{
		Default: true,
		MD001:   &RuleConfig{Enabled: boolPtr(true)},
		MD003:   &RuleConfig{Enabled: boolPtr(true)},
		MD009:   &RuleConfig{Enabled: boolPtr(true)},
		MD010:   &RuleConfig{Enabled: boolPtr(true)},
		MD012:   &RuleConfig{Enabled: boolPtr(true)},
		MD013:   &RuleConfig{Enabled: boolPtr(true)},
		MD018:   &RuleConfig{Enabled: boolPtr(true)},
		MD019:   &RuleConfig{Enabled: boolPtr(true)},
		MD023:   &RuleConfig{Enabled: boolPtr(true)},
		MD032:   &RuleConfig{Enabled: boolPtr(true)},
		MD047:   &RuleConfig{Enabled: boolPtr(true)},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

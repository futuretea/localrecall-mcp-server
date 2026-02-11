package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// StaticConfig represents the static configuration for the LocalRecall MCP Server
type StaticConfig struct {
	// Server configuration
	Port       int    `mapstructure:"port"`
	SSEBaseURL string `mapstructure:"sse_base_url"`

	// Logging configuration
	LogLevel int `mapstructure:"log_level"`

	// LocalRecall configuration
	LocalRecallURL        string `mapstructure:"localrecall_url"`
	LocalRecallAPIKey     string `mapstructure:"localrecall_api_key"`
	LocalRecallCollection string `mapstructure:"localrecall_collection"`

	// Output configuration
	ListOutput    string   `mapstructure:"list_output"`
	OutputFilters []string `mapstructure:"output_filters"`

	// Tool configuration
	EnabledTools  []string `mapstructure:"enabled_tools"`
	DisabledTools []string `mapstructure:"disabled_tools"`
}

// Validate validates the configuration
func (c *StaticConfig) Validate() error {
	// Validate port
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535, got %d", c.Port)
	}

	// Validate log level
	if c.LogLevel < 0 || c.LogLevel > 9 {
		return fmt.Errorf("log_level must be between 0 and 9, got %d", c.LogLevel)
	}

	// Validate list output
	validOutputs := map[string]bool{
		"table": true,
		"yaml":  true,
		"json":  true,
	}
	if !validOutputs[strings.ToLower(c.ListOutput)] {
		return fmt.Errorf("list_output must be one of: table, yaml, json, got %s", c.ListOutput)
	}

	// Validate LocalRecall URL
	if c.LocalRecallURL != "" {
		if !strings.HasPrefix(c.LocalRecallURL, "http://") && !strings.HasPrefix(c.LocalRecallURL, "https://") {
			return fmt.Errorf("localrecall_url must start with http:// or https://, got %s", c.LocalRecallURL)
		}
	}

	return nil
}

// LoadConfig loads configuration from file and environment variables using Viper
// Priority: command-line flags > environment variables > config file > defaults
func LoadConfig(configPath string) (*StaticConfig, error) {
	// Use the global viper instance to access bound command-line flags
	v := viper.GetViper()

	// Set defaults
	v.SetDefault("port", 0)
	v.SetDefault("log_level", 5)
	v.SetDefault("localrecall_url", "http://localhost:8080")
	v.SetDefault("list_output", "json")

	// Set configuration file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Configure environment variable support
	// Environment variables use LOCALRECALL_MCP_ prefix and replace - with _
	v.SetEnvPrefix("LOCALRECALL_MCP")
	v.AllowEmptyEnv(true)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// Unmarshal configuration into struct
	config := &StaticConfig{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// HasLocalRecallConfig returns true if LocalRecall configuration is present
func (c *StaticConfig) HasLocalRecallConfig() bool {
	return c.LocalRecallURL != ""
}

// GetPortString returns the port as a string in the format ":port"
func (c *StaticConfig) GetPortString() string {
	if c.Port == 0 {
		return ""
	}
	return fmt.Sprintf(":%d", c.Port)
}

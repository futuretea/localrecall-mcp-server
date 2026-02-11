package handler

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// FormatJSON formats data as JSON
func FormatJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// FormatYAML formats data as YAML
func FormatYAML(data interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(yamlBytes), nil
}

// FormatOutput formats data according to the specified format
func FormatOutput(data interface{}, format string) (string, error) {
	switch format {
	case "yaml":
		return FormatYAML(data)
	case "json":
		return FormatJSON(data)
	default:
		return FormatJSON(data)
	}
}

package handler

import (
	"fmt"
)

// GetStringParam extracts a string parameter from the params map
func GetStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetIntParam extracts an integer parameter from the params map
func GetIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			// Try to parse string as int
			var intVal int
			if _, err := fmt.Sscanf(v, "%d", &intVal); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

// RequireStringParam extracts a required string parameter
func RequireStringParam(params map[string]interface{}, key string) (string, error) {
	val, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	if strVal == "" {
		return "", fmt.Errorf("parameter %s cannot be empty", key)
	}
	return strVal, nil
}

package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// formatAsEnv formats the data as shell export statements
func formatAsEnv(data map[string]string) string {
	if len(data) == 0 {
		return ""
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		value := data[key]
		// Escape single quotes in value
		escapedValue := strings.ReplaceAll(value, "'", "'\\''")
		lines = append(lines, fmt.Sprintf("export %s='%s'", key, escapedValue))
	}

	return strings.Join(lines, "\n")
}

// formatAsJSON formats the data as JSON
func formatAsJSON(data map[string]string) (string, error) {
	if len(data) == 0 {
		return "{}", nil
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// formatAsYAML formats the data as YAML
func formatAsYAML(data map[string]string) (string, error) {
	if len(data) == 0 {
		return "{}", nil
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// formatAsDotEnv formats the data as .env file format
func formatAsDotEnv(data map[string]string) string {
	if len(data) == 0 {
		return ""
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		value := data[key]
		// Quote value if it contains spaces or special characters
		if strings.ContainsAny(value, " \t\n\"'$") {
			// Escape quotes and use double quotes
			escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
			lines = append(lines, fmt.Sprintf("%s=\"%s\"", key, escapedValue))
		} else {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return strings.Join(lines, "\n")
}

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/furiatona/azctl/internal/logx"
)

// fetchAzureAppConfig queries Azure App Configuration via az CLI and returns key-value pairs.
// It expects 'az appconfig kv list' to be available; if not, returns empty map.
//
//nolint:unused // This function is used as a wrapper for fetchAzureAppConfigWithImage
func fetchAzureAppConfig(ctx context.Context, name, label string) (map[string]string, error) {
	return fetchAzureAppConfigWithImage(ctx, name, label, "")
}

// fetchAzureAppConfigWithImage queries Azure App Configuration with image name support
func fetchAzureAppConfigWithImage(ctx context.Context, name, label, imageName string) (map[string]string, error) {
	if name == "" {
		return map[string]string{}, nil
	}

	logx.Infof("[DEBUG] Fetching from Azure App Config: name='%s', label='%s'", name, label)

	// Initialize result map
	m := map[string]string{}

	// First, try to get the global-configurations key specifically
	logx.Infof("[DEBUG] Trying to fetch global-configurations key specifically")
	globalArgs := []string{"appconfig", "kv", "show", "--name", name, "--key", "global-configurations",
		"--query", "{key:key,value:value}", "-o", "json"}
	if label != "" {
		globalArgs = append(globalArgs, "--label", label)
	}
	globalCmd := exec.CommandContext(ctx, "az", globalArgs...) //nolint:gosec // az cli is trusted
	globalOut, globalErr := globalCmd.Output()

	if globalErr == nil {
		logx.Infof("[DEBUG] Found global-configurations key: %s", string(globalOut))
		var globalKV struct{ Key, Value string }
		if err := json.Unmarshal(globalOut, &globalKV); err == nil {
			// Parse the JSON value from global-configurations
			var globalConfig map[string]interface{}
			if err := json.Unmarshal([]byte(globalKV.Value), &globalConfig); err == nil {
				for k, v := range globalConfig {
					if str, ok := v.(string); ok {
						keyName := strings.ToUpper(k)
						logx.Infof("[DEBUG] Adding from global-configurations: %s='%s'", keyName, str)
						m[keyName] = str
					}
				}
			} else {
				//nolint:errcheck // Error logging for debugging
				logx.Errorf("[ERROR] Failed to parse global-configurations JSON: %v", err)
				//nolint:errcheck // Error logging for debugging
				logx.Errorf("[ERROR] Raw JSON value: %s", globalKV.Value)
				return nil, fmt.Errorf("malformed JSON in Azure App Configuration global-configurations key: %w", err)
			}
		}
	} else {
		logx.Infof("[DEBUG] global-configurations key not found or error: %v", globalErr)
		// Try without label if label was specified
		if label != "" {
			logx.Infof("[DEBUG] Trying global-configurations without label")
			globalArgsNoLabel := []string{"appconfig", "kv", "show", "--name", name, "--key", "global-configurations",
				"--query", "{key:key,value:value}", "-o", "json"}
			globalCmdNoLabel := exec.CommandContext(ctx, "az", globalArgsNoLabel...) //nolint:gosec // az cli is trusted
			globalOutNoLabel, globalErrNoLabel := globalCmdNoLabel.Output()

			if globalErrNoLabel == nil {
				logx.Infof("[DEBUG] Found global-configurations key without label: %s", string(globalOutNoLabel))
				var globalKVNoLabel struct{ Key, Value string }
				if err := json.Unmarshal(globalOutNoLabel, &globalKVNoLabel); err == nil {
					// Parse the JSON value from global-configurations
					var globalConfigNoLabel map[string]interface{}
					if err := json.Unmarshal([]byte(globalKVNoLabel.Value), &globalConfigNoLabel); err == nil {
						for k, v := range globalConfigNoLabel {
							if str, ok := v.(string); ok {
								keyName := strings.ToUpper(k)
								logx.Infof("[DEBUG] Adding from global-configurations (no label): %s='%s'", keyName, str)
								m[keyName] = str
							}
						}
					} else {
						//nolint:errcheck // Error logging for debugging
						logx.Errorf("[ERROR] Failed to parse global-configurations JSON (no label): %v", err)
						//nolint:errcheck // Error logging for debugging
						logx.Errorf("[ERROR] Raw JSON value (no label): %s", globalKVNoLabel.Value)
						return nil, fmt.Errorf("malformed JSON in Azure App Configuration global-configurations key (no label): %w", err)
					}
				}
			} else {
				logx.Infof("[DEBUG] global-configurations key not found without label either: %v", globalErrNoLabel)
			}
		}
	}

	// Second, try to get the service-specific key (e.g., swarm-embedding-service)
	if imageName != "" {
		logx.Infof("[DEBUG] Trying to fetch service-specific key: '%s'", imageName)
		serviceArgs := []string{"appconfig", "kv", "show", "--name", name, "--key", imageName,
			"--query", "{key:key,value:value}", "-o", "json"}
		if label != "" {
			serviceArgs = append(serviceArgs, "--label", label)
		}
		serviceCmd := exec.CommandContext(ctx, "az", serviceArgs...) //nolint:gosec // az cli is trusted
		serviceOut, serviceErr := serviceCmd.Output()

		if serviceErr == nil {
			logx.Infof("[DEBUG] Found service-specific key: %s", string(serviceOut))
			var serviceKV struct{ Key, Value string }
			if err := json.Unmarshal(serviceOut, &serviceKV); err == nil {
				// Parse the JSON value from service-specific key
				var serviceConfig map[string]interface{}
				if err := json.Unmarshal([]byte(serviceKV.Value), &serviceConfig); err == nil {
					for k, v := range serviceConfig {
						if str, ok := v.(string); ok {
							logx.Infof("[DEBUG] Adding from service-specific key: %s='%s'", strings.ToUpper(k), str)
							m[strings.ToUpper(k)] = str
						}
					}
				} else {
					//nolint:errcheck // Error logging for debugging
					logx.Errorf("[ERROR] Failed to parse service-specific JSON for key '%s': %v", imageName, err)
					//nolint:errcheck // Error logging for debugging
					logx.Errorf("[ERROR] Please check your Azure App Configuration JSON format for key '%s'", imageName)
					//nolint:errcheck // Error logging for debugging
					logx.Errorf("[ERROR] Common issues: missing commas, duplicate keys, or invalid JSON syntax")
					//nolint:errcheck // Error logging for debugging
					logx.Errorf("[ERROR] Raw service JSON value: %s", serviceKV.Value)
					return nil, fmt.Errorf("malformed JSON in Azure App Configuration key '%s': %w", imageName, err)
				}
			}
		} else {
			logx.Infof("[DEBUG] Service-specific key '%s' not found or error: %v", imageName, serviceErr)
			// Try without label if label was specified
			if label != "" {
				logx.Infof("[DEBUG] Trying service-specific key without label")
				serviceArgsNoLabel := []string{"appconfig", "kv", "show", "--name", name, "--key", imageName,
					"--query", "{key:key,value:value}", "-o", "json"}
				serviceCmdNoLabel := exec.CommandContext(ctx, "az", serviceArgsNoLabel...) //nolint:gosec // az cli is trusted
				serviceOutNoLabel, serviceErrNoLabel := serviceCmdNoLabel.Output()

				if serviceErrNoLabel == nil {
					logx.Infof("[DEBUG] Found service-specific key without label: %s", string(serviceOutNoLabel))
					var serviceKVNoLabel struct{ Key, Value string }
					if err := json.Unmarshal(serviceOutNoLabel, &serviceKVNoLabel); err == nil {
						// Parse the JSON value from service-specific key
						var serviceConfigNoLabel map[string]interface{}
						if err := json.Unmarshal([]byte(serviceKVNoLabel.Value), &serviceConfigNoLabel); err == nil {
							for k, v := range serviceConfigNoLabel {
								if str, ok := v.(string); ok {
									logx.Infof("[DEBUG] Adding from service-specific key (no label): %s='%s'", strings.ToUpper(k), str)
									m[strings.ToUpper(k)] = str
								}
							}
						} else {
							//nolint:errcheck // Error logging for debugging
							logx.Errorf("[ERROR] Failed to parse service-specific JSON (no label) for key '%s': %v", imageName, err)
							//nolint:errcheck // Error logging for debugging
							logx.Errorf("[ERROR] Please check your Azure App Configuration JSON format for key '%s'", imageName)
							//nolint:errcheck // Error logging for debugging
							logx.Errorf("[ERROR] Common issues: missing commas, duplicate keys, or invalid JSON syntax")
							//nolint:errcheck // Error logging for debugging
							logx.Errorf("[ERROR] Raw service JSON value (no label): %s", serviceKVNoLabel.Value)
							return nil, fmt.Errorf("malformed JSON in Azure App Configuration key '%s': %w", imageName, err)
						}
					}
				} else {
					logx.Infof("[DEBUG] Service-specific key '%s' not found without label either: %v", imageName, serviceErrNoLabel)
				}
			}
		}
	}

	// Return the results we have so far (global-configurations + service-specific)
	logx.Infof("[DEBUG] Returning config with %d variables", len(m))
	return m, nil
}

// ExportAllConfig exports all configuration from Azure App Configuration
func ExportAllConfig(ctx context.Context, name, label string) (map[string]string, error) {
	if name == "" {
		return nil, fmt.Errorf("APP_CONFIG_NAME is required")
	}

	logx.Infof("[DEBUG] Exporting all config from: name='%s', label='%s'", name, label)

	// Build az appconfig kv list command
	args := []string{"appconfig", "kv", "list", "--name", name, "-o", "json"}
	if label != "" {
		args = append(args, "--label", label)
	}

	cmd := exec.CommandContext(ctx, "az", args...) //nolint:gosec // az cli is trusted
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list app config keys: %w", err)
	}

	// Parse the output
	var kvList []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(out, &kvList); err != nil {
		return nil, fmt.Errorf("failed to parse app config output: %w", err)
	}

	result := make(map[string]string)
	for _, kv := range kvList {
		// Check if value is JSON (for global-configurations and service keys)
		var jsonValue map[string]interface{}
		if err := json.Unmarshal([]byte(kv.Value), &jsonValue); err == nil {
			// It's a JSON object, extract key-value pairs
			for k, v := range jsonValue {
				if str, ok := v.(string); ok {
					result[strings.ToUpper(k)] = str
				}
			}
		} else {
			// It's a plain value, use key as-is
			result[strings.ToUpper(kv.Key)] = kv.Value
		}
	}

	logx.Infof("[DEBUG] Exported %d variables", len(result))
	return result, nil
}

// ExportSpecificVars exports specific variables from Azure App Configuration
func ExportSpecificVars(ctx context.Context, name, label string, vars []string) (map[string]string, error) {
	if name == "" {
		return nil, fmt.Errorf("APP_CONFIG_NAME is required")
	}

	if len(vars) == 0 {
		return make(map[string]string), nil
	}

	logx.Infof("[DEBUG] Exporting specific vars from: name='%s', label='%s', vars=%v", name, label, vars)

	// First, get all config
	allConfig, err := ExportAllConfig(ctx, name, label)
	if err != nil {
		return nil, err
	}

	// Filter to only requested variables
	result := make(map[string]string)
	for _, varName := range vars {
		upperVar := strings.ToUpper(varName)
		if value, ok := allConfig[upperVar]; ok {
			result[upperVar] = value
		} else {
			logx.Infof("[WARNING] Variable '%s' not found in app config", varName)
		}
	}

	logx.Infof("[DEBUG] Exported %d of %d requested variables", len(result), len(vars))
	return result, nil
}

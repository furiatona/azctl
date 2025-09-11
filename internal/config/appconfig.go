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

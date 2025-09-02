package config

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/furiatona/azctl/internal/logx"
)

// fetchAzureAppConfig queries Azure App Configuration via az CLI and returns key-value pairs.
// It expects 'az appconfig kv list' to be available; if not, returns empty map.
// nolint:unused // This function is used as a wrapper for fetchAzureAppConfigWithImage
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
	globalArgs := []string{"appconfig", "kv", "show", "--name", name, "--key", "global-configurations", "--query", "{key:key,value:value}", "-o", "json"}
	globalCmd := exec.CommandContext(ctx, "az", globalArgs...)
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
						// Map ACR_REGISTRY to REGISTRY for compatibility
						keyName := strings.ToUpper(k)
						if keyName == "ACR_REGISTRY" {
							keyName = "REGISTRY"
						}
						logx.Infof("[DEBUG] Adding from global-configurations: %s='%s'", keyName, str)
						m[keyName] = str
					}
				}
			} else {
				logx.Infof("[DEBUG] Failed to parse global-configurations JSON: %v", err)
				logx.Infof("[DEBUG] Raw JSON value: %s", globalKV.Value)
			}
		}
	} else {
		logx.Infof("[DEBUG] global-configurations key not found or error: %v", globalErr)
	}

	// Second, try to get the service-specific key (e.g., swarm-embedding-service)
	if imageName != "" {
		logx.Infof("[DEBUG] Trying to fetch service-specific key: '%s'", imageName)
		serviceArgs := []string{"appconfig", "kv", "show", "--name", name, "--key", imageName, "--query", "{key:key,value:value}", "-o", "json"}
		serviceCmd := exec.CommandContext(ctx, "az", serviceArgs...)
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
					logx.Infof("[DEBUG] Failed to parse service-specific JSON: %v", err)
					logx.Infof("[DEBUG] Raw service JSON value: %s", serviceKV.Value)
				}
			}
		} else {
			logx.Infof("[DEBUG] Service-specific key '%s' not found or error: %v", imageName, serviceErr)
		}
	}

	// Return the results we have so far (global-configurations + service-specific)
	logx.Infof("[DEBUG] Returning config with %d variables", len(m))
	return m, nil
}

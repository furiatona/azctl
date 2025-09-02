package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"
)

// reloadConfigWithEnvironment reloads configuration with environment-specific Azure App Configuration
func reloadConfigWithEnvironment(ctx context.Context, envName string) error {
	// Get current config to check for environment-specific app config names
	cfg := config.Current()

	// Look for environment-specific app config name (e.g., DEV_APP_CONFIG)
	envAppConfigKey := fmt.Sprintf("%s_APP_CONFIG", strings.ToUpper(envName))
	appConfigName := cfg.Get(envAppConfigKey)

	logx.Infof("[DEBUG] Looking for %s in config: '%s'", envAppConfigKey, appConfigName)

	// Fallback to generic format if not found
	if appConfigName == "" {
		appConfigName = fmt.Sprintf("app-config-%s", envName)
		logx.Infof("[DEBUG] Using fallback app config name: '%s'", appConfigName)
	} else {
		logx.Infof("[DEBUG] Using environment-specific app config: '%s'", appConfigName)
	}

	appConfigLabel := envName

	// Temporarily set these environment variables
	originalAppConfigName := os.Getenv("APP_CONFIG_NAME")
	originalAppConfigLabel := os.Getenv("APP_CONFIG_LABEL")

	logx.Infof("[DEBUG] Setting APP_CONFIG_NAME='%s', APP_CONFIG_LABEL='%s'", appConfigName, appConfigLabel)

	// nolint:errcheck // os.Setenv rarely fails in this context
	os.Setenv("APP_CONFIG_NAME", appConfigName)
	os.Setenv("APP_CONFIG_LABEL", appConfigLabel) // nolint:errcheck

	// Reload configuration with .env file path
	envfile := ".env" // Default .env file path
	logx.Infof("[DEBUG] Reloading config with envfile: %s", envfile)
	if err := config.Init(ctx, envfile); err != nil {
		// Restore original values on error
		if originalAppConfigName != "" {
			// nolint:errcheck // os.Setenv rarely fails in error recovery
			os.Setenv("APP_CONFIG_NAME", originalAppConfigName)
		} else {
			// nolint:errcheck // os.Unsetenv rarely fails in error recovery
			os.Unsetenv("APP_CONFIG_NAME")
		}
		if originalAppConfigLabel != "" {
			// nolint:errcheck // os.Setenv rarely fails in error recovery
			os.Setenv("APP_CONFIG_LABEL", originalAppConfigLabel)
		} else {
			// nolint:errcheck // os.Unsetenv rarely fails in error recovery
			os.Unsetenv("APP_CONFIG_LABEL")
		}
		return err
	}

	return nil
}

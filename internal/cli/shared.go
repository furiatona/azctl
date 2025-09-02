package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"
)

// isCIEnvironment detects if we're running in a CI environment
func isCIEnvironment() bool {
	// Check for common CI environment variables
	ciVars := []string{"CI", "GITHUB_ACTIONS", "AZURE_PIPELINE", "GITLAB_CI", "JENKINS_URL", "TRAVIS", "CIRCLECI"}
	for _, envVar := range ciVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}

// detectEnvironmentFromCI detects the current environment from CI context
func detectEnvironmentFromCI() string {
	// Try to detect from GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		ref := os.Getenv("GITHUB_REF")
		if strings.HasPrefix(ref, "refs/heads/") {
			branch := strings.TrimPrefix(ref, "refs/heads/")
			switch branch {
			case "dev", "development":
				return "dev"
			case "staging":
				return "staging"
			case "main", "master", "prod", "production":
				return "prod"
			}
		}
	}

	// Try to detect from Azure Pipeline
	if os.Getenv("AZURE_PIPELINE") == "true" {
		// Azure Pipeline environment variables
		if env := os.Getenv("SYSTEM_ENVIRONMENT"); env != "" {
			return strings.ToLower(env)
		}
	}

	// Try to detect from GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		if env := os.Getenv("CI_ENVIRONMENT_NAME"); env != "" {
			return strings.ToLower(env)
		}
	}

	// Try to detect from explicit environment variable
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return strings.ToLower(env)
	}

	return ""
}

// mapEnvironmentVariables maps environment-specific prefixed variables to non-prefixed versions in CI
func mapEnvironmentVariables(cfg *config.Config, envName string) {
	if !isCIEnvironment() {
		return // Only map in CI environment
	}

	if envName == "" {
		envName = detectEnvironmentFromCI()
		if envName == "" {
			logx.Infof("[DEBUG] Could not detect environment in CI, skipping variable mapping")
			return
		}
	}

	logx.Infof("[DEBUG] Mapping environment variables for CI environment: %s", envName)

	// Auto-detect IMAGE_NAME and IMAGE_TAG if not set
	if cfg.Get("IMAGE_NAME") == "" {
		if detectedImageName := detectImageNameFromCI(); detectedImageName != "" {
			cfg.Set("IMAGE_NAME", detectedImageName)
			logx.Infof("[DEBUG] Auto-detected IMAGE_NAME: %s", detectedImageName)
		}
	}

	if cfg.Get("IMAGE_TAG") == "" {
		if detectedImageTag := detectImageTagFromCI(); detectedImageTag != "" {
			cfg.Set("IMAGE_TAG", detectedImageTag)
			logx.Infof("[DEBUG] Auto-detected IMAGE_TAG: %s", detectedImageTag)
		}
	}

	// Define the mapping of prefixed variables to non-prefixed versions
	envPrefix := strings.ToUpper(envName) + "_"

	// Common variable mappings
	variableMappings := []string{
		"RESOURCE_GROUP",
		"APP_CONFIG",
		"APP_CONFIG_RESOURCE_GROUP",
		"WEBAPP_NAME",
		"APP_SERVICE_PLAN",
		"DNS_NAME_LABEL",
		"ACI_PORT",
		"ACI_CPU",
		"ACI_MEMORY",
		"OS_TYPE",
		"LOCATION",
		"REGISTRY",
		"ACR_RESOURCE_GROUP",
		"CONTAINER_GROUP_NAME",
		"IMAGE_NAME",
		"IMAGE_TAG",
		"ACR_USERNAME",
		"ACR_PASSWORD",
		"LOG_SHARE_NAME",
		"LOG_STORAGE_ACCOUNT",
		"LOG_STORAGE_KEY",
		"FLUENTBIT_CONFIG_SHARE",
		"APP_CONFIG_SKIP",
		"APP_CONFIG_DEBUG",
	}

	// Map each variable
	for _, varName := range variableMappings {
		prefixedVar := envPrefix + varName
		prefixedValue := cfg.Get(prefixedVar)

		if prefixedValue != "" {
			// Check if non-prefixed version is already set
			nonPrefixedValue := cfg.Get(varName)
			if nonPrefixedValue == "" {
				// Map prefixed to non-prefixed
				cfg.Set(varName, prefixedValue)
				logx.Infof("[DEBUG] Mapped %s='%s' to %s", prefixedVar, prefixedValue, varName)
			} else {
				logx.Infof("[DEBUG] Non-prefixed %s already set to '%s', keeping it", varName, nonPrefixedValue)
			}
		}
	}

	// Special handling for ACI_* variables
	aciMappings := map[string]string{
		"ACI_SUPABASE_KEY":                         "SUPABASE_KEY",
		"ACI_SUPABASE_URL":                         "SUPABASE_URL",
		"ACI_AZURE_OPENAI_API_KEY":                 "AZURE_OPENAI_API_KEY",
		"ACI_OPENAI_AZURE_EMBEDDINGS_ENDPOINT":     "OPENAI_AZURE_EMBEDDINGS_ENDPOINT",
		"ACI_AZURE_OPENAI_MODEL":                   "AZURE_OPENAI_MODEL",
		"ACI_FIREBASE_KEY":                         "FIREBASE_KEY",
		"ACI_FIREBASE_URL":                         "FIREBASE_URL",
		"ACI_SAGEMAKER_OPENAI_MODEL":               "SAGEMAKER_OPENAI_MODEL",
		"ACI_SAGEMAKER_OPENAI_API_KEY":             "SAGEMAKER_OPENAI_API_KEY",
		"ACI_OPENAI_SAGEMAKER_EMBEDDINGS_ENDPOINT": "OPENAI_SAGEMAKER_EMBEDDINGS_ENDPOINT",
	}

	for aciKey, mappedKey := range aciMappings {
		prefixedAciKey := envPrefix + aciKey
		prefixedValue := cfg.Get(prefixedAciKey)

		if prefixedValue != "" {
			// Check if non-prefixed version is already set
			nonPrefixedValue := cfg.Get(mappedKey)
			if nonPrefixedValue == "" {
				// Map prefixed to non-prefixed
				cfg.Set(mappedKey, prefixedValue)
				logx.Infof("[DEBUG] Mapped %s='%s' to %s", prefixedAciKey, prefixedValue, mappedKey)
			} else {
				logx.Infof("[DEBUG] Non-prefixed %s already set to '%s', keeping it", mappedKey, nonPrefixedValue)
			}
		}
	}
}

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

	// Map environment variables if in CI
	if isCIEnvironment() {
		mapEnvironmentVariables(config.Current(), envName)
	}

	return nil
}

// detectImageNameFromCI detects the image name from CI context
func detectImageNameFromCI() string {
	// Try to detect from GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		if repoName := os.Getenv("GITHUB_REPOSITORY"); repoName != "" {
			// Extract repository name from "owner/repo" format
			parts := strings.Split(repoName, "/")
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	// Try to detect from Azure Pipeline
	if os.Getenv("AZURE_PIPELINE") == "true" {
		if buildRepoName := os.Getenv("BUILD_REPOSITORY_NAME"); buildRepoName != "" {
			return buildRepoName
		}
	}

	// Try to detect from GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		if projectName := os.Getenv("CI_PROJECT_NAME"); projectName != "" {
			return projectName
		}
	}

	return ""
}

// detectImageTagFromCI detects the image tag from CI context
func detectImageTagFromCI() string {
	// Try to detect from GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		if sha := os.Getenv("GITHUB_SHA"); sha != "" {
			return sha
		}
	}

	// Try to detect from Azure Pipeline
	if os.Getenv("AZURE_PIPELINE") == "true" {
		if buildId := os.Getenv("BUILD_BUILDID"); buildId != "" {
			return buildId
		}
		if sourceVersion := os.Getenv("BUILD_SOURCEVERSION"); sourceVersion != "" {
			return sourceVersion
		}
	}

	// Try to detect from GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		if commitSha := os.Getenv("CI_COMMIT_SHA"); commitSha != "" {
			return commitSha
		}
	}

	return ""
}

package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logging"
	"github.com/furiatona/azctl/internal/runx"
	"github.com/furiatona/azctl/internal/validation"

	"github.com/spf13/cobra"
)

func newWebAppCmd() *cobra.Command {
	var (
		resourceGroup  string
		webAppName     string
		appServicePlan string
	)

	cmd := &cobra.Command{
		Use:   "webapp",
		Short: "Deploy to Azure Web App using container image from ACR",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get environment from root command
			envName, _ := cmd.Flags().GetString("env")

			cfg := config.Current()

			// Auto-detect environment in CI if not provided
			if envName == "" && isCIEnvironment() {
				detectedEnv := detectEnvironmentFromCI()
				if detectedEnv != "" {
					envName = detectedEnv
					logging.Debugf("Auto-detected environment in CI: %s", envName)
				}
			}

			// Environment is required for WebApp deployment
			if envName == "" {
				return fmt.Errorf("environment required for webapp deployment (--env dev|staging|prod)")
			}

			// Auto-detect IMAGE_NAME and IMAGE_TAG in CI if not set
			if isCIEnvironment() {
				if cfg.Get("IMAGE_NAME") == "" {
					if detectedImageName := detectImageNameFromCI(); detectedImageName != "" {
						cfg.Set("IMAGE_NAME", detectedImageName)
						logging.Debugf("Auto-detected IMAGE_NAME from CI: %s", detectedImageName)
					}
				}
				if cfg.Get("IMAGE_TAG") == "" {
					if detectedImageTag := detectImageTagFromCI(); detectedImageTag != "" {
						cfg.Set("IMAGE_TAG", detectedImageTag)
						logging.Debugf("Auto-detected IMAGE_TAG from CI: %s", detectedImageTag)
					}
				}
			}

			// Apply flag overrides
			if resourceGroup == "" {
				resourceGroup = cfg.Get("RESOURCE_GROUP")
			}
			if webAppName == "" {
				webAppName = getWebAppName(cfg, envName)
			}
			if appServicePlan == "" {
				appServicePlan = getAppServicePlan(cfg, envName)
			}

			// Validate required variables
			if err := validation.RequiredVars(cfg, validation.WebAppRequiredVars()); err != nil {
				return fmt.Errorf("WebApp deployment validation failed: %w", err)
			}

			// Check if WebApp exists
			exists, err := checkWebAppExists(cmd.Context(), resourceGroup, webAppName)
			if err != nil {
				return fmt.Errorf("failed to check WebApp existence: %w", err)
			}

			if exists {
				// Update existing WebApp
				logging.Infof("Updating existing Web App '%s'...", webAppName)
				return updateWebApp(cmd.Context(), resourceGroup, webAppName, cfg)
			} else {
				// Create new WebApp
				if appServicePlan == "" {
					return fmt.Errorf("WebApp '%s' does not exist and APP_SERVICE_PLAN not provided. "+
						"Please either:\n1. Set APP_SERVICE_PLAN environment variable to create new web apps, or\n"+
						"2. Create the web app manually first, or\n"+
						"3. Use a different web app name that already exists", webAppName)
				}

				logging.Infof("Creating new Web App '%s'...", webAppName)
				if err := createWebApp(cmd.Context(), resourceGroup, webAppName, appServicePlan); err != nil {
					return fmt.Errorf("failed to create WebApp: %w", err)
				}
				return updateWebApp(cmd.Context(), resourceGroup, webAppName, cfg)
			}
		},
	}

	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group (env: RESOURCE_GROUP)")
	cmd.Flags().StringVar(&webAppName, "name", "", "WebApp name (env: WEBAPP_NAME or <env>_WEBAPP_NAME)")
	cmd.Flags().StringVar(&appServicePlan, "plan", "",
		"App Service Plan (env: APP_SERVICE_PLAN or <env>_APP_SERVICE_PLAN)")
	return cmd
}

// getWebAppName determines the WebApp name based on environment and configuration
func getWebAppName(cfg *config.Config, env string) string {
	// Check for environment-specific name first
	envSpecificKey := fmt.Sprintf("%s_WEBAPP_NAME", strings.ToUpper(env))
	if name := cfg.Get(envSpecificKey); name != "" {
		return name
	}

	// Check for generic WEBAPP_NAME
	if name := cfg.Get("WEBAPP_NAME"); name != "" {
		return name
	}

	// Default naming convention
	imageName := cfg.Get("IMAGE_NAME")
	if imageName == "" {
		imageName = "webapp"
	}
	return fmt.Sprintf("%s-%s", imageName, env)
}

// getAppServicePlan determines the App Service Plan based on environment and configuration
func getAppServicePlan(cfg *config.Config, env string) string {
	// Check for environment-specific plan first
	envSpecificKey := fmt.Sprintf("%s_APP_SERVICE_PLAN", strings.ToUpper(env))
	if plan := cfg.Get(envSpecificKey); plan != "" {
		return plan
	}

	// Check for generic APP_SERVICE_PLAN
	return cfg.Get("APP_SERVICE_PLAN")
}

// checkWebAppExists checks if a WebApp exists in the specified resource group
func checkWebAppExists(ctx context.Context, resourceGroup, webAppName string) (bool, error) {
	args := []string{"webapp", "show", "--name", webAppName, "--resource-group", resourceGroup}
	err := runx.AZ(ctx, args...)
	return err == nil, nil // If command succeeds, WebApp exists
}

// createWebApp creates a new WebApp
func createWebApp(ctx context.Context, resourceGroup, webAppName, appServicePlan string) error {
	args := []string{
		"webapp", "create",
		"--resource-group", resourceGroup,
		"--plan", appServicePlan,
		"--name", webAppName,
	}
	if err := runx.AZ(ctx, args...); err != nil {
		return fmt.Errorf("failed to create webapp: %w", err)
	}
	return nil
}

// updateWebApp updates an existing WebApp with container configuration
func updateWebApp(ctx context.Context, resourceGroup, webAppName string, cfg *config.Config) error {
	registry := cfg.Get("ACR_REGISTRY")
	imageName := cfg.Get("IMAGE_NAME")
	imageTag := cfg.Get("IMAGE_TAG")

	if registry == "" || imageName == "" || imageTag == "" {
		return fmt.Errorf("missing required variables: ACR_REGISTRY, IMAGE_NAME, IMAGE_TAG")
	}

	fullImageName := fmt.Sprintf("%s/%s:%s", registry, imageName, imageTag)
	registryUrl := fmt.Sprintf("https://%s", registry)

	// Set container image
	args := []string{
		"webapp", "config", "container", "set",
		"--name", webAppName,
		"--resource-group", resourceGroup,
		"--container-image-name", fullImageName,
		"--container-registry-url", registryUrl,
	}
	if err := runx.AZ(ctx, args...); err != nil {
		return fmt.Errorf("failed to update webapp container: %w", err)
	}

	// Set application settings (environment variables) from config
	if err := setWebAppSettings(ctx, resourceGroup, webAppName, cfg); err != nil {
		return fmt.Errorf("failed to set webapp settings: %w", err)
	}

	return nil
}

// setWebAppSettings sets application settings (environment variables) for the WebApp
func setWebAppSettings(ctx context.Context, resourceGroup, webAppName string, cfg *config.Config) error {
	// Collect only application-specific environment variables (like ACI does)
	allVars := cfg.GetAll()
	settings := make([]string, 0, len(allVars))
	for key, value := range allVars {
		// Skip internal azctl variables that shouldn't be passed to the container
		if isInternalVariable(key) {
			continue
		}

		// Skip variables with very long values that might cause Azure CLI issues
		if len(value) > 4000 {
			logging.Debugf("Skipping variable '%s' - value too long (%d chars)", key, len(value))
			continue
		}

		// Only include variables that are application-specific (similar to ACI environmentVariables)
		if !isApplicationVariable(key) {
			logging.Debugf("Skipping infrastructure variable '%s'", key)
			continue
		}

		// Escape the value for shell safety (but don't add quotes)
		escapedValue := escapeShellValue(value)
		settings = append(settings, fmt.Sprintf("%s=%s", key, escapedValue))
		logging.Debugf("Including application setting: %s", key)
	}

	if len(settings) == 0 {
		logging.Debugf("No application settings to configure for WebApp '%s'", webAppName)
		return nil
	}

	// Set application settings using az CLI - do it in batches to avoid command line length limits
	const batchSize = 20
	for i := 0; i < len(settings); i += batchSize {
		end := i + batchSize
		if end > len(settings) {
			end = len(settings)
		}

		batch := settings[i:end]
		args := []string{
			"webapp", "config", "appsettings", "set",
			"--name", webAppName,
			"--resource-group", resourceGroup,
			"--settings",
		}
		args = append(args, batch...)

		logging.Debugf("Setting batch %d/%d (%d settings) for WebApp '%s'",
			(i/batchSize)+1, (len(settings)+batchSize-1)/batchSize, len(batch), webAppName)

		if err := runx.AZ(ctx, args...); err != nil {
			return fmt.Errorf("failed to set application settings batch %d: %w", (i/batchSize)+1, err)
		}
	}

	logging.Infof("✅ Set %d application settings for WebApp '%s'", len(settings), webAppName)
	return nil
}

// escapeShellValue escapes a value for safe use in shell commands
func escapeShellValue(value string) string {
	// Replace quotes with escaped quotes and handle special characters
	// Don't wrap in quotes as Azure CLI handles the values properly
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	return escaped
}

// isInternalVariable checks if a variable is internal to azctl and shouldn't be passed to containers
func isInternalVariable(key string) bool {
	internalVars := []string{
		"ACR_REGISTRY",
		"ACR_RESOURCE_GROUP",
		"ACR_USERNAME",
		"ACR_PASSWORD",
		"RESOURCE_GROUP",
		"IMAGE_NAME",
		"IMAGE_TAG",
		"WEBAPP_NAME",
		"APP_SERVICE_PLAN",
		"LOG_STORAGE_ACCOUNT",
		"LOG_STORAGE_KEY",
		"LOG_STORAGE_NAME",
		"FLUENTBIT_CONFIG",
		"APP_CONFIG_NAME",
		"APP_CONFIG_LABEL",
		"APP_CONFIG_SKIP",
	}

	for _, internal := range internalVars {
		if key == internal {
			return true
		}
	}
	return false
}

// isApplicationVariable checks if a variable should be passed to the application container
// This matches the environmentVariables in the ACI template
func isApplicationVariable(key string) bool {
	// Application-specific prefixes and variables (like in ACI environmentVariables)
	applicationPrefixes := []string{
		"NEXT_PUBLIC_",
		"SUPABASE_",
		"SOLANA_",
		"AZURE_OPENAI_",
		"OPENAI_",
		"LOGFLARE_",
		"FIREBASE_",
		"SAGEMAKER_",
	}

	// Check prefixes
	for _, prefix := range applicationPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Specific application variables (not prefixed)
	applicationVars := []string{
		"PORT",
		"NODE_ENV",
		"ENVIRONMENT",
	}

	for _, appVar := range applicationVars {
		if key == appVar {
			return true
		}
	}

	return false
}

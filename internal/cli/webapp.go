package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"
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
					logx.Infof("[DEBUG] Auto-detected environment in CI: %s", envName)
				}
			}

			// Environment is required for WebApp deployment
			if envName == "" {
				return fmt.Errorf("environment required for webapp deployment (--env dev|staging|prod)")
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
				logx.Infof("Updating existing Web App '%s'...", webAppName)
				return updateWebApp(cmd.Context(), resourceGroup, webAppName, cfg)
			} else {
				// Create new WebApp
				if appServicePlan == "" {
					return fmt.Errorf("WebApp '%s' does not exist and APP_SERVICE_PLAN not provided. Please either:\n1. Set APP_SERVICE_PLAN environment variable to create new web apps, or\n2. Create the web app manually first, or\n3. Use a different web app name that already exists", webAppName)
				}

				logx.Infof("Creating new Web App '%s'...", webAppName)
				if err := createWebApp(cmd.Context(), resourceGroup, webAppName, appServicePlan); err != nil {
					return fmt.Errorf("failed to create WebApp: %w", err)
				}
				return updateWebApp(cmd.Context(), resourceGroup, webAppName, cfg)
			}
		},
	}

	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group (env: RESOURCE_GROUP)")
	cmd.Flags().StringVar(&webAppName, "name", "", "WebApp name (env: WEBAPP_NAME or <env>_WEBAPP_NAME)")
	cmd.Flags().StringVar(&appServicePlan, "plan", "", "App Service Plan (env: APP_SERVICE_PLAN or <env>_APP_SERVICE_PLAN)")
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
	return runx.AZ(ctx, args...)
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

	args := []string{
		"webapp", "config", "container", "set",
		"--name", webAppName,
		"--resource-group", resourceGroup,
		"--container-image-name", fullImageName,
		"--container-registry-url", registryUrl,
	}
	return runx.AZ(ctx, args...)
}

package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"
	"github.com/furiatona/azctl/internal/runx"

	"github.com/spf13/cobra"
)

// findACRResourceGroup finds the resource group containing the specified ACR
func findACRResourceGroup(ctx context.Context, registryName string) string {
	if registryName == "" {
		return ""
	}

	// List all resource groups and check for ACR
	args := []string{"group", "list", "--query", "[].name", "-o", "tsv"}
	output, err := runx.AZOutput(ctx, args...)
	if err != nil {
		logx.Warnf("Failed to list resource groups: %v", err)
		return ""
	}

	resourceGroups := strings.Split(strings.TrimSpace(output), "\n")
	for _, rg := range resourceGroups {
		rg = strings.TrimSpace(rg)
		if rg == "" {
			continue
		}

		// Check if ACR exists in this resource group
		checkArgs := []string{"acr", "show", "--name", registryName, "--resource-group", rg}
		if _, err := runx.AZOutput(ctx, checkArgs...); err == nil {
			return rg
		}
	}

	return ""
}

// collectBuildArgs collects NEXT_PUBLIC_* variables for build arguments
func collectBuildArgs(cfg *config.Config) []string {
	var buildArgs []string
	for key, value := range cfg.GetAll() {
		if strings.HasPrefix(key, "NEXT_PUBLIC_") {
			buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=%s", key, value))
			logx.Infof("[DEBUG] Adding build arg: %s='%s'", key, value)
		}
	}
	return buildArgs
}

func newACRCmd() *cobra.Command {
	var (
		envName       string
		registry      string
		resourceGroup string
		imageName     string
		imageTag      string
		contextPath   string
		file          string
	)

	cmd := &cobra.Command{
		Use:   "acr",
		Short: "Build and push Docker image to Azure Container Registry",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := config.Current()

			// Auto-detect environment in CI if not provided
			if envName == "" && isCIEnvironment() {
				detectedEnv := detectEnvironmentFromCI()
				if detectedEnv != "" {
					envName = detectedEnv
					logx.Infof("[DEBUG] Auto-detected environment in CI: %s", envName)
				}
			}

			// If environment is specified, reload config with environment-specific Azure App Configuration
			if envName != "" {
				logx.Infof("[DEBUG] Loading environment-specific config for: %s", envName)
				if err := reloadConfigWithEnvironment(cmd.Context(), envName); err != nil {
					return fmt.Errorf("failed to load environment config: %w", err)
				}
				cfg = config.Current() // Get the updated config
				logx.Infof("[DEBUG] Config reloaded for environment: %s", envName)
			}

			// Apply flag overrides to config
			if registry != "" {
				cfg.Set("REGISTRY", registry)
			}
			if resourceGroup != "" {
				cfg.Set("ACR_RESOURCE_GROUP", resourceGroup)
			}
			if imageName != "" {
				cfg.Set("IMAGE_NAME", imageName)
			}
			if imageTag != "" {
				cfg.Set("IMAGE_TAG", imageTag)
			}

			// Validate required variables
			requiredVars := []string{"IMAGE_NAME", "IMAGE_TAG", "REGISTRY"}
			for _, varName := range requiredVars {
				if cfg.Get(varName) == "" {
					return fmt.Errorf("missing required variable: %s", varName)
				}
			}

			// Get ACR resource group
			acrResourceGroup := cfg.Get("ACR_RESOURCE_GROUP")
			if acrResourceGroup == "" {
				// Try to find ACR in any resource group
				acrResourceGroup = findACRResourceGroup(cmd.Context(), cfg.Get("REGISTRY"))
				if acrResourceGroup == "" {
					return fmt.Errorf("ACR resource group not found for registry: %s", cfg.Get("REGISTRY"))
				}
				logx.Infof("Found ACR in resource group: %s", acrResourceGroup)
			}

			// Build and push image
			imageName = cfg.Get("IMAGE_NAME")
			imageTag = cfg.Get("IMAGE_TAG")
			registry = cfg.Get("REGISTRY")
			fullImageName := fmt.Sprintf("%s.azurecr.io/%s:%s", registry, imageName, imageTag)

			logx.Printf("Building and pushing image: %s", fullImageName)

			// Use az acr build command
			args := []string{
				"acr", "build",
				"--registry", registry,
				"--image", fmt.Sprintf("%s:%s", imageName, imageTag),
				"--resource-group", acrResourceGroup,
			}

			// Add Dockerfile path if specified
			if file != "" {
				args = append(args, "--file", file)
			}

			// Add build arguments if any NEXT_PUBLIC_ variables are set
			buildArgs := collectBuildArgs(cfg)
			if len(buildArgs) > 0 {
				logx.Infof("Adding build arguments: %v", buildArgs)
				args = append(args, buildArgs...)
			}

			// Add context path (defaults to ".")
			if contextPath == "" {
				contextPath = "."
			}
			args = append(args, contextPath)

			if err := runx.AZ(cmd.Context(), args...); err != nil {
				return fmt.Errorf("failed to build and push image: %w", err)
			}

			logx.Printf("Successfully built and pushed image: %s", fullImageName)
			return nil
		},
	}

	cmd.Flags().StringVar(&envName, "env", "", "Environment name; optional to select app config scope (auto-detected in CI)")
	cmd.Flags().StringVar(&registry, "registry", "", "ACR registry name (env: REGISTRY)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group for ACR (env: ACR_RESOURCE_GROUP)")
	cmd.Flags().StringVar(&imageName, "image", "", "Image name (env: IMAGE_NAME)")
	cmd.Flags().StringVar(&imageTag, "tag", "", "Image tag (env: IMAGE_TAG)")
	cmd.Flags().StringVar(&contextPath, "context", ".", "Build context path")
	cmd.Flags().StringVar(&file, "file", "", "Dockerfile path")
	return cmd
}

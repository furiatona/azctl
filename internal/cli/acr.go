package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logging"
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
		logging.Warnf("Failed to list resource groups: %v", err)
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
			logging.Debugf("Adding build arg: %s='%s'", key, value)
		}
	}
	return buildArgs
}

func newACRCmd() *cobra.Command {
	var (
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

			// Apply flag overrides to config
			if registry != "" {
				cfg.Set("ACR_REGISTRY", registry)
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

			// Validate required variables
			requiredVars := []string{"IMAGE_NAME", "IMAGE_TAG"}
			for _, varName := range requiredVars {
				if cfg.Get(varName) == "" {
					return fmt.Errorf("missing required variable: %s", varName)
				}
			}

			// Check for registry (ACR_REGISTRY)
			registry = cfg.Get("ACR_REGISTRY")
			if registry == "" {
				return fmt.Errorf("missing required variable: ACR_REGISTRY")
			}

			// Get ACR resource group
			acrResourceGroup := cfg.Get("ACR_RESOURCE_GROUP")
			if acrResourceGroup == "" {
				// Try to find ACR in any resource group
				acrResourceGroup = findACRResourceGroup(cmd.Context(), registry)
				if acrResourceGroup == "" {
					return fmt.Errorf("ACR resource group not found for registry: %s", registry)
				}
				logging.Infof("Found ACR in resource group: %s", acrResourceGroup)
			}

			// Build and push image
			imageName = cfg.Get("IMAGE_NAME")
			imageTag = cfg.Get("IMAGE_TAG")
			fullImageName := fmt.Sprintf("%s.azurecr.io/%s:%s", registry, imageName, imageTag)

			// Check if image already exists
			logging.Infof("Checking if image already exists: %s", fullImageName)
			checkArgs := []string{
				"acr", "repository", "show-tags",
				"--name", registry,
				"--repository", imageName,
				"--output", "tsv",
			}
			existingTags, err := runx.AZOutput(cmd.Context(), checkArgs...)
			if err == nil {
				// Check if the tag exists
				if strings.Contains(existingTags, imageTag) {
					logging.Infof("✅ Image already exists: %s", fullImageName)
					logging.Infof("Skipping build for existing image")
					return nil
				}
			}

			logging.Infof("Building and pushing image: %s", fullImageName)

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
				logging.Debugf("Adding build arguments: %v", buildArgs)
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

			logging.Infof("Successfully built and pushed image: %s", fullImageName)
			return nil
		},
	}

	cmd.Flags().StringVar(&registry, "registry", "", "ACR registry name (env: ACR_REGISTRY)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group for ACR (env: ACR_RESOURCE_GROUP)")
	cmd.Flags().StringVar(&imageName, "image", "", "Image name (env: IMAGE_NAME)")
	cmd.Flags().StringVar(&imageTag, "tag", "", "Image tag (env: IMAGE_TAG)")
	cmd.Flags().StringVar(&contextPath, "context", ".", "Build context path")
	cmd.Flags().StringVar(&file, "file", "", "Dockerfile path")
	return cmd
}

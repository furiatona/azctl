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

func newACRCmd() *cobra.Command {
	var (
		registry      string
		resourceGroup string
		imageName     string
		imageTag      string
		contextPath   string
		file          string
		envName       string
	)

	cmd := &cobra.Command{
		Use:   "acr",
		Short: "Build and push image to Azure Container Registry via az acr build",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Current()

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
				// Flag values take highest precedence, but we still use cfg.Get for consistency
			} else {
				registry = cfg.Get("REGISTRY")
			}
			if resourceGroup != "" {
				// Flag values take highest precedence
			} else {
				resourceGroup = cfg.Get("ACR_RESOURCE_GROUP")
			}
			if imageName != "" {
				// Flag values take highest precedence
			} else {
				imageName = cfg.Get("IMAGE_NAME")
			}
			if imageTag != "" {
				// Flag values take highest precedence
			} else {
				imageTag = cfg.Get("IMAGE_TAG")
			}

			// Debug: Show resolved values
			logx.Infof("[DEBUG] Resolved values - REGISTRY: '%s', ACR_RESOURCE_GROUP: '%s', IMAGE_NAME: '%s', IMAGE_TAG: '%s'",
				registry, resourceGroup, imageName, imageTag)

			// Collect NEXT_PUBLIC_* variables for build arguments
			var buildArgs []string
			for key, value := range cfg.GetAll() {
				if strings.HasPrefix(key, "NEXT_PUBLIC_") {
					buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=%s", key, value))
					logx.Infof("[DEBUG] Adding build arg: %s='%s'", key, value)
				}
			}
			logx.Infof("[DEBUG] Found %d NEXT_PUBLIC_* build arguments", len(buildArgs)/2)

			// Create a temporary config with resolved values for validation
			tempCfg := &config.Config{}
			tempCfg.Set("REGISTRY", registry)
			tempCfg.Set("ACR_RESOURCE_GROUP", resourceGroup)
			tempCfg.Set("IMAGE_NAME", imageName)
			tempCfg.Set("IMAGE_TAG", imageTag)

			if err := validation.RequiredVars(tempCfg, validation.ACRRequiredVars()); err != nil {
				return err
			}

			// Check if image already exists to avoid unnecessary rebuilds
			imageExists, err := checkImageExists(cmd.Context(), registry, imageName, imageTag)
			if err != nil {
				logx.Warnf("Failed to check if image exists: %v", err)
			} else if imageExists {
				logx.Printf("✅ Image %s:%s already exists in registry %s, skipping build", imageName, imageTag, registry)
				return nil
			}

			target := fmt.Sprintf("%s.azurecr.io/%s:%s", registry, imageName, imageTag)
			azArgs := []string{
				"acr", "build",
				"--resource-group", resourceGroup,
				"--registry", registry,
				"--image", target,
			}
			if file != "" {
				azArgs = append(azArgs, "--file", file)
			}
			// Add build arguments for NEXT_PUBLIC_* variables
			azArgs = append(azArgs, buildArgs...)
			if contextPath == "" {
				contextPath = "."
			}
			azArgs = append(azArgs, contextPath)

			return runx.AZ(cmd.Context(), azArgs...)
		},
	}

	cmd.Flags().StringVar(&registry, "registry", "", "ACR registry name (env: REGISTRY)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group for ACR (env: ACR_RESOURCE_GROUP)")
	cmd.Flags().StringVar(&imageName, "image", "", "Image name (env: IMAGE_NAME)")
	cmd.Flags().StringVar(&imageTag, "tag", "", "Image tag (env: IMAGE_TAG)")
	cmd.Flags().StringVar(&contextPath, "context", ".", "Build context path")
	cmd.Flags().StringVar(&file, "file", "", "Dockerfile path")
	cmd.Flags().StringVar(&envName, "env", "", "Environment name; optional to select app config scope")
	return cmd
}

// checkImageExists checks if an image with the specified name and tag already exists in the ACR
func checkImageExists(ctx context.Context, registry, imageName, imageTag string) (bool, error) {
	logx.Infof("[DEBUG] Checking if image %s:%s exists in registry %s", imageName, imageTag, registry)

	// Use az acr repository show-tags to get all tags for the repository
	args := []string{
		"acr", "repository", "show-tags",
		"--name", registry,
		"--repository", imageName,
		"--output", "json",
	}

	output, err := runx.AZOutput(ctx, args...)
	if err != nil {
		logx.Infof("[DEBUG] Image existence check failed: %v", err)
		// If the command fails, it might mean the image doesn't exist
		return false, nil
	}

	logx.Infof("[DEBUG] Image existence check output: %s", output)

	// Check if the tag exists in the repository
	exists := strings.Contains(output, imageTag)
	if exists {
		logx.Infof("[DEBUG] Image %s:%s found in registry", imageName, imageTag)
	} else {
		logx.Infof("[DEBUG] Image %s:%s not found in registry", imageName, imageTag)
	}

	return exists, nil
}

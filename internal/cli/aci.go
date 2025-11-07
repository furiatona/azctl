package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logging"
	"github.com/furiatona/azctl/internal/runx"
	"github.com/furiatona/azctl/internal/templatex"
	"github.com/furiatona/azctl/internal/validation"

	"github.com/spf13/cobra"
)

const (
	envProd        = "prod"
	envProduction  = "production"
	envDev         = "dev"
	envDevelopment = "development"
	envStaging     = "staging"
	envTrue        = "true"
)

func newACICmd() *cobra.Command {
	var (
		resourceGroup string
		templatePath  string
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "aci",
		Short: "Deploy Azure Container Instance with sidecar using JSON template",
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

			// Set environment name for Fluent-bit configuration
			if envName != "" {
				cfg.Set("ENV_NAME", envName)
				logging.Debugf("Set ENV_NAME='%s' for Fluent-bit config", envName)
			}

			// Auto-detect IMAGE_NAME, IMAGE_TAG, CONTAINER_GROUP_NAME, and DNS_NAME_LABEL in CI if not set
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
				if cfg.Get("CONTAINER_GROUP_NAME") == "" {
					if detectedImageName := detectImageNameFromCI(); detectedImageName != "" {
						cfg.Set("CONTAINER_GROUP_NAME", detectedImageName)
						logging.Debugf("Auto-detected CONTAINER_GROUP_NAME from CI: %s", detectedImageName)
					}
				}
				if cfg.Get("DNS_NAME_LABEL") == "" {
					containerName := cfg.Get("CONTAINER_GROUP_NAME")
					if containerName != "" && envName != "" {
						dnsNameLabel := fmt.Sprintf("%s-%s", containerName, envName)
						cfg.Set("DNS_NAME_LABEL", dnsNameLabel)
						logging.Debugf("Auto-detected DNS_NAME_LABEL from CI: %s", dnsNameLabel)
					}
				}
			}

			if templatePath == "" {
				templatePath = "deploy/manifests/aci.json"
			}
			if _, err := os.Stat(templatePath); err != nil {
				// fallback to local azctl/aci.json if user provided reference in repo
				if _, err2 := os.Stat("azctl/aci.json"); err2 == nil {
					templatePath = "azctl/aci.json"
				} else {
					return fmt.Errorf("template not found: %s", templatePath)
				}
			}

			// Apply flag overrides
			if resourceGroup == "" {
				resourceGroup = cfg.Get("RESOURCE_GROUP")
			}

			// Map environment-specific resource groups to RESOURCE_GROUP
			if resourceGroup == "" {
				envResourceGroupKey := fmt.Sprintf("%s_RESOURCE_GROUP", strings.ToUpper(envName))
				resourceGroup = cfg.Get(envResourceGroupKey)
				if resourceGroup != "" {
					cfg.Set("RESOURCE_GROUP", resourceGroup)
					logging.Debugf("Mapped %s='%s' to RESOURCE_GROUP", envResourceGroupKey, resourceGroup)
				}
			}

			// Map ACR_REGISTRY to IMAGE_REGISTRY for template compatibility
			if cfg.Get("IMAGE_REGISTRY") == "" {
				acrRegistry := cfg.Get("ACR_REGISTRY")
				if acrRegistry != "" {
					cfg.Set("IMAGE_REGISTRY", acrRegistry)
					logging.Debugf("Mapped ACR_REGISTRY='%s' to IMAGE_REGISTRY", acrRegistry)
				}
			}

			// Set environment-based defaults if not provided
			applyACIDefaults(cfg, envName)

			// Validate all required ACI variables are present
			if err := validation.RequiredVars(cfg, validation.ACIRequiredVars()); err != nil {
				return fmt.Errorf("ACI deployment validation failed: %w", err)
			}

			// render template by replacing {{VAR}} placeholders with values from cfg
			raw, err := os.ReadFile(templatePath) //nolint:gosec // templatePath is validated
			if err != nil {
				return fmt.Errorf("failed to read template file: %w", err)
			}
			rendered, err := templatex.RenderEnv(string(raw), cfg)
			if err != nil {
				return fmt.Errorf("failed to render template: %w", err)
			}

			// validate JSON
			var js map[string]any
			if err := json.Unmarshal([]byte(rendered), &js); err != nil {
				return fmt.Errorf("rendered JSON invalid: %w", err)
			}

			// Generate Fluent-bit configuration for logging integration
			loggingManager := logging.NewManager()
			if err := loggingManager.GenerateConfig(cfg, cfg.Get("IMAGE_NAME"), envName); err != nil {
				return fmt.Errorf("failed to generate logging config: %w", err)
			}

			if dryRun {
				// Create .azctl directory if it doesn't exist
				if err := os.MkdirAll(".azctl", 0755); err != nil { //nolint:gosec // acceptable permissions for directory
					return fmt.Errorf("failed to create .azctl directory: %w", err)
				}

				// Write rendered JSON to .azctl/aci-dry-run.json
				outputFile := ".azctl/aci-dry-run.json"
				if err := os.WriteFile(outputFile, []byte(rendered), 0644); err != nil {
					return fmt.Errorf("failed to write dry-run output: %w", err)
				}

				logging.Infof("Dry run complete. Generated ACI JSON written to: %s", outputFile)
				logging.Infof("Review the file and run without --dry-run to deploy")
				return nil
			}

			// Handle different deployment strategies based on environment
			if err := deployACI(cmd.Context(), resourceGroup, envName, rendered); err != nil {
				return fmt.Errorf("ACI deployment failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group (env: AZURE_RESOURCE_GROUP)")
	cmd.Flags().StringVar(&templatePath, "template", "", "Path to aci.json template")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Generate ACI JSON without deploying (outputs to .azctl/aci-dry-run.json)")
	return cmd
}

// applyACIDefaults sets reasonable defaults for ACI deployment if not already configured
func applyACIDefaults(cfg *config.Config, envName string) {
	defaults := map[string]string{
		"LOCATION":   "eastus",
		"OS_TYPE":    "Linux",
		"ACI_PORT":   "8080",
		"ACI_CPU":    "1",
		"ACI_MEMORY": "2",
		// Storage defaults for logging sidecar
		"LOG_SHARE_NAME":         "applogs",
		"LOG_STORAGE_ACCOUNT":    "swarmlogs",
		"LOG_STORAGE_KEY":        "placeholder-key",
		"FLUENTBIT_CONFIG_SHARE": "fluentbit-config",
	}

	// Apply defaults only if values are not already set
	for key, defaultValue := range defaults {
		if cfg.Get(key) == "" {
			cfg.Set(key, defaultValue)
		}
	}

	// Environment-specific defaults
	if envName != "" {
		// Set DNS_NAME_LABEL if not provided
		if cfg.Get("DNS_NAME_LABEL") == "" {
			containerName := cfg.Get("CONTAINER_GROUP_NAME")
			if containerName != "" {
				cfg.Set("DNS_NAME_LABEL", fmt.Sprintf("%s-%s", containerName, envName))
			} else {
				// Use IMAGE_NAME as fallback for container group name
				imageName := cfg.Get("IMAGE_NAME")
				if imageName != "" {
					cfg.Set("CONTAINER_GROUP_NAME", imageName)
					cfg.Set("DNS_NAME_LABEL", fmt.Sprintf("%s-%s", imageName, envName))
				}
			}
		}
	}
}

// deployACI handles different deployment strategies based on environment
func deployACI(ctx context.Context, resourceGroup, envName, rendered string) error {
	// For dev and staging: check if container group exists, delete it, then create new one
	if envName == "dev" || envName == "development" || envName == "staging" {
		cfg := config.Current()
		containerGroupName := cfg.Get("CONTAINER_GROUP_NAME")
		if containerGroupName == "" {
			containerGroupName = cfg.Get("IMAGE_NAME") // fallback to image name
		}

		logging.Infof("🔍 Environment: %s - Checking for existing container group: %s", envName, containerGroupName)

		// Check if container group exists
		exists, err := checkContainerGroupExists(ctx, resourceGroup, containerGroupName)
		if err != nil {
			logging.Warnf("Failed to check if container group exists: %v", err)
		} else if exists {
			logging.Infof("🗑️  Container group %s exists. Deleting it...", containerGroupName)
			if err := deleteContainerGroup(ctx, resourceGroup, containerGroupName); err != nil {
				return fmt.Errorf("failed to delete existing container group: %w", err)
			}
			logging.Infof("✅ Container group %s deleted successfully", containerGroupName)
		} else {
			logging.Infof("📝 Container group %s does not exist. Proceeding with creation...", containerGroupName)
		}
	}

	// Create new container group
	logging.Infof("🚀 Creating new container group...")
	return createContainerGroup(ctx, resourceGroup, rendered)
}

// checkContainerGroupExists checks if a container group exists in the specified resource group
func checkContainerGroupExists(ctx context.Context, resourceGroup, containerGroupName string) (bool, error) {
	args := []string{
		"container", "show",
		"--resource-group", resourceGroup,
		"--name", containerGroupName,
		"--output", "json",
	}

	_, err := runx.AZOutput(ctx, args...)
	if err != nil {
		// If the command fails, the container group likely doesn't exist
		return false, nil
	}
	return true, nil
}

// deleteContainerGroup deletes an existing container group
func deleteContainerGroup(ctx context.Context, resourceGroup, containerGroupName string) error {
	args := []string{
		"container", "delete",
		"--resource-group", resourceGroup,
		"--name", containerGroupName,
		"--yes", // Skip confirmation
	}

	if err := runx.AZ(ctx, args...); err != nil {
		return fmt.Errorf("failed to delete container group: %w", err)
	}
	return nil
}

// createContainerGroup creates a new container group from JSON
func createContainerGroup(ctx context.Context, resourceGroup, rendered string) error {
	// Write to temp file for az cli
	f, err := os.CreateTemp("", "aci-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			logging.Warnf("failed to remove temp file %s: %v", f.Name(), err)
		}
	}()
	if _, err := f.WriteString(rendered); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := runx.AZ(ctx, "container", "create", "--resource-group", resourceGroup, "--file", f.Name()); err != nil {
		return fmt.Errorf("failed to create container group: %w", err)
	}
	return nil
}

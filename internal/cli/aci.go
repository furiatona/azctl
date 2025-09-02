package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logging"
	"github.com/furiatona/azctl/internal/logx"
	"github.com/furiatona/azctl/internal/runx"
	"github.com/furiatona/azctl/internal/templatex"
	"github.com/furiatona/azctl/internal/validation"

	"github.com/spf13/cobra"
)

func newACICmd() *cobra.Command {
	var (
		resourceGroup string
		envName       string
		templatePath  string
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "aci",
		Short: "Deploy Azure Container Instance with sidecar using JSON template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Check for production environment early
			if envName == "prod" || envName == "production" {
				logx.Infof("Production deployment is coming soon!")
				logx.Infof("For now, please use --dry-run to generate the ACI JSON and deploy manually.")
				return nil
			}

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
				resourceGroup = cfg.Get("AZURE_RESOURCE_GROUP")
			}

			// Map environment-specific resource groups to AZURE_RESOURCE_GROUP
			if resourceGroup == "" {
				envResourceGroupKey := fmt.Sprintf("%s_RESOURCE_GROUP", strings.ToUpper(envName))
				resourceGroup = cfg.Get(envResourceGroupKey)
				if resourceGroup != "" {
					cfg.Set("AZURE_RESOURCE_GROUP", resourceGroup)
					logx.Infof("[DEBUG] Mapped %s='%s' to AZURE_RESOURCE_GROUP", envResourceGroupKey, resourceGroup)
				}
			}

			// Map REGISTRY to IMAGE_REGISTRY for ACI compatibility
			if cfg.Get("IMAGE_REGISTRY") == "" {
				registry := cfg.Get("REGISTRY")
				if registry != "" {
					cfg.Set("IMAGE_REGISTRY", registry)
					logx.Infof("[DEBUG] Mapped REGISTRY='%s' to IMAGE_REGISTRY", registry)
				}
			}

			// Map ACI_* variables to their non-prefixed versions for template compatibility
			aciMappings := map[string]string{
				"ACI_SUPABASE_KEY":                     "SUPABASE_KEY",
				"ACI_SUPABASE_URL":                     "SUPABASE_URL",
				"ACI_AZURE_OPENAI_API_KEY":             "AZURE_OPENAI_API_KEY",
				"ACI_OPENAI_AZURE_EMBEDDINGS_ENDPOINT": "OPENAI_AZURE_EMBEDDINGS_ENDPOINT",
				"ACI_AZURE_OPENAI_MODEL":               "AZURE_OPENAI_MODEL",
			}

			for aciKey, mappedKey := range aciMappings {
				if cfg.Get(mappedKey) == "" {
					aciValue := cfg.Get(aciKey)
					if aciValue != "" {
						cfg.Set(mappedKey, aciValue)
						logx.Infof("[DEBUG] Mapped %s='%s' to %s", aciKey, aciValue, mappedKey)
					}
				}
			}

			// Set environment name for Fluent-bit configuration
			if envName != "" {
				cfg.Set("ENV_NAME", envName)
				logx.Infof("[DEBUG] Set ENV_NAME='%s' for Fluent-bit config", envName)
			}

			// Set environment-based defaults if not provided
			applyACIDefaults(cfg, envName)

			// Validate all required ACI variables are present
			if err := validation.RequiredVars(cfg, validation.ACIRequiredVars()); err != nil {
				return fmt.Errorf("ACI deployment validation failed: %w", err)
			}

			// render template by replacing {{VAR}} placeholders with values from cfg
			raw, err := os.ReadFile(templatePath)
			if err != nil {
				return err
			}
			rendered, err := templatex.RenderEnv(string(raw), cfg)
			if err != nil {
				return err
			}

			// validate JSON
			var js map[string]any
			if err := json.Unmarshal([]byte(rendered), &js); err != nil {
				return fmt.Errorf("rendered JSON invalid: %w", err)
			}

			// Generate Fluent-bit configuration for logging integration
			loggingManager := logging.NewManager()
			if err := loggingManager.GenerateConfig(cfg, cfg.Get("IMAGE_NAME"), envName); err != nil {
				logx.Warnf("Failed to generate logging config: %v", err)
			}

			if dryRun {
				// Create .azctl directory if it doesn't exist
				if err := os.MkdirAll(".azctl", 0755); err != nil {
					return fmt.Errorf("failed to create .azctl directory: %w", err)
				}

				// Write rendered JSON to .azctl/aci-dry-run.json
				outputFile := ".azctl/aci-dry-run.json"
				if err := os.WriteFile(outputFile, []byte(rendered), 0644); err != nil {
					return fmt.Errorf("failed to write dry-run output: %w", err)
				}

				logx.Infof("Dry run complete. Generated ACI JSON written to: %s", outputFile)
				logx.Infof("Review the file and run without --dry-run to deploy")
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
	cmd.Flags().StringVar(&envName, "env", "", "Environment name; optional to select app config scope (auto-detected in CI)")
	cmd.Flags().StringVar(&templatePath, "template", "", "Path to aci.json template")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Generate ACI JSON without deploying (outputs to .azctl/aci-dry-run.json)")
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

		logx.Printf("🔍 Environment: %s - Checking for existing container group: %s", envName, containerGroupName)

		// Check if container group exists
		exists, err := checkContainerGroupExists(ctx, resourceGroup, containerGroupName)
		if err != nil {
			logx.Warnf("Failed to check if container group exists: %v", err)
		} else if exists {
			logx.Printf("🗑️  Container group %s exists. Deleting it...", containerGroupName)
			if err := deleteContainerGroup(ctx, resourceGroup, containerGroupName); err != nil {
				return fmt.Errorf("failed to delete existing container group: %w", err)
			}
			logx.Printf("✅ Container group %s deleted successfully", containerGroupName)
		} else {
			logx.Printf("📝 Container group %s does not exist. Proceeding with creation...", containerGroupName)
		}
	}

	// Create new container group
	logx.Printf("🚀 Creating new container group...")
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

	return runx.AZ(ctx, args...)
}

// createContainerGroup creates a new container group from JSON
func createContainerGroup(ctx context.Context, resourceGroup, rendered string) error {
	// Write to temp file for az cli
	f, err := os.CreateTemp("", "aci-*.json")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			logx.Warnf("failed to remove temp file %s: %v", f.Name(), err)
		}
	}()
	if _, err := f.WriteString(rendered); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return runx.AZ(ctx, "container", "create", "--resource-group", resourceGroup, "--file", f.Name())
}

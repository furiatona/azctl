package cli

import (
	"fmt"
	"os"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logging"

	"github.com/spf13/cobra"
)

func newAppConfigCmd() *cobra.Command {
	var (
		vars       []string
		format     string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "appconfig",
		Short: "Export configuration from Azure App Configuration",
		Long: `Export configuration variables from Azure App Configuration.
		
By default, exports all configuration variables. Use --var to export specific variables.
Supports multiple output formats: env, json, yaml, dotenv.

Examples:
  # Export all variables in env format
  azctl appconfig --env dev

  # Export specific variables
  azctl appconfig --env prod --var DATABASE_URL --var REDIS_URL

  # Export to file in JSON format
  azctl appconfig --env staging --format json --output config.json

  # Export as dotenv file
  azctl appconfig --env dev --format dotenv --output .env.exported`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get environment from root command
			envName, _ := cmd.Flags().GetString("env")

			cfg := config.Current()

			// Get APP_CONFIG_NAME
			appConfigName := cfg.Get("APP_CONFIG_NAME")
			if appConfigName == "" {
				// Try APP_CONFIG as fallback
				appConfigName = cfg.Get("APP_CONFIG")
			}
			if appConfigName == "" {
				return fmt.Errorf("APP_CONFIG_NAME or APP_CONFIG environment variable is required")
			}

			// Determine label from environment
			label := envName
			if envLabel := cfg.Get("APP_CONFIG_LABEL"); envLabel != "" {
				label = envLabel
			}

			logging.Infof("Exporting from Azure App Configuration: %s (label: %s)", appConfigName, label)

			// Export configuration
			var data map[string]string
			var err error

			if len(vars) > 0 {
				// Export specific variables
				logging.Infof("Exporting specific variables: %v", vars)
				data, err = config.ExportSpecificVars(cmd.Context(), appConfigName, label, vars)
			} else {
				// Export all variables
				logging.Infof("Exporting all variables")
				data, err = config.ExportAllConfig(cmd.Context(), appConfigName, label)
			}

			if err != nil {
				return fmt.Errorf("failed to export config: %w", err)
			}

			if len(data) == 0 {
				logging.Warnf("No configuration variables found")
				return nil
			}

			logging.Infof("Exported %d variable(s)", len(data))

			// Format output
			var output string
			switch format {
			case "env":
				output = formatAsEnv(data)
			case "json":
				output, err = formatAsJSON(data)
				if err != nil {
					return fmt.Errorf("failed to format as JSON: %w", err)
				}
			case "yaml":
				output, err = formatAsYAML(data)
				if err != nil {
					return fmt.Errorf("failed to format as YAML: %w", err)
				}
			case "dotenv":
				output = formatAsDotEnv(data)
			default:
				return fmt.Errorf("unsupported format: %s (supported: env, json, yaml, dotenv)", format)
			}

			// Write output
			if outputFile != "" {
				// Write to file
				if err := os.WriteFile(outputFile, []byte(output), 0600); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
				logging.Infof("Configuration exported to: %s", outputFile)
			} else {
				// Write to stdout
				fmt.Println(output)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&vars, "var", nil, "Specific variable(s) to export (can be specified multiple times)")
	cmd.Flags().StringVar(&format, "format", "env", "Output format: env, json, yaml, dotenv")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file (default: stdout)")

	return cmd
}


package cli

import (
	"context"
	"fmt"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"

	"github.com/spf13/cobra"
)

func Execute(ctx context.Context, args []string) error {
	root := &cobra.Command{
		Use:           "azctl",
		Short:         "Azure CLI wrapper and deployment helper",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global persistent flags
	root.PersistentFlags().String("envfile", ".env", "Path to .env file (optional)")
	root.PersistentFlags().String("env", "", "Environment name (dev, staging, prod) - determines .env file and Azure App Config scope")
	root.PersistentFlags().Bool("verbose", false, "Enable verbose logging")

	// Initialize config/logging before running any subcommand
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		logx.Init(verbose)

		envfile, _ := cmd.Flags().GetString("envfile")
		env, _ := cmd.Flags().GetString("env")
		
		// If environment is specified, use environment-specific .env file
		if env != "" && envfile == ".env" {
			envfile = fmt.Sprintf(".env.%s", env)
		}
		
		if err := config.Init(cmd.Context(), envfile, env); err != nil {
			return fmt.Errorf("init config: %w", err)
		}
		return nil
	}

	// Subcommands
	root.AddCommand(newACRCmd())
	root.AddCommand(newACICmd())
	root.AddCommand(newWebAppCmd())

	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

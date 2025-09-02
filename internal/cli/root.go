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
	root.PersistentFlags().Bool("verbose", false, "Enable verbose logging")

	// Initialize config/logging before running any subcommand
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		logx.Init(verbose)

		envfile, _ := cmd.Flags().GetString("envfile")
		if err := config.Init(cmd.Context(), envfile); err != nil {
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

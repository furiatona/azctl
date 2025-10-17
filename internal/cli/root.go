package cli

import (
	"context"
	"fmt"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logging"
	"github.com/furiatona/azctl/internal/logx"

	"github.com/spf13/cobra"
)

// Version information - imported from main package
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// SetVersionInfo sets the version information from main package
func SetVersionInfo(version, buildTime, gitCommit string) {
	Version = version
	BuildTime = buildTime
	GitCommit = gitCommit
}

func Execute(ctx context.Context, args []string) error {
	root := &cobra.Command{
		Use:   "azctl",
		Short: "Professional Azure Container Deployment CLI Tool",
		Long: `A production-ready Go CLI tool that provides a seamless interface for Azure container deployment workflows. ` +
			`Simplifies container management and deployment to Azure services.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (build: %s, commit: %s)", Version, BuildTime, GitCommit),
	}

	// Global persistent flags
	root.PersistentFlags().String("envfile", ".env", "Path to .env file (optional)")
	root.PersistentFlags().String("env", "",
		"Environment name (dev, staging, prod) - determines .env file and Azure App Config scope")
	root.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	root.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	root.PersistentFlags().String("log-format", "text", "Log format (text, json)")

	// Initialize config/logging before running any subcommand
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		logLevel, _ := cmd.Flags().GetString("log-level")
		logFormat, _ := cmd.Flags().GetString("log-format")

		// Initialize logging
		logConfig := logging.Config{
			Level:     logging.Level(logLevel),
			Formatter: logging.Formatter(logFormat),
		}
		if verbose {
			logConfig.Level = logging.LevelDebug
		}
		if err := logging.Init(logConfig); err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}

		// Initialize logx package with verbose flag for Azure App Configuration logging
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
	root.AddCommand(newAppConfigCmd())

	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}

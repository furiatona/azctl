package main

import (
	"context"
	"fmt"
	"os"

	"github.com/furiatona/azctl/internal/cli"
)

// Version information - set during build
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Set version information in CLI package
	cli.SetVersionInfo(version, buildTime, gitCommit)

	if err := cli.Execute(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

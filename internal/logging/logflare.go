package logging

import (
	"fmt"
	"os"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/templatex"
)

// LogflareProvider implements LoggingProvider for Logflare
type LogflareProvider struct{}

func (p *LogflareProvider) Name() string {
	return "Logflare"
}

func (p *LogflareProvider) IsEnabled(cfg *config.Config) bool {
	logflareAPIKey := cfg.Get("LOGFLARE_API_KEY")
	logflareSourceID := cfg.Get("LOGFLARE_SOURCE_ID")
	return logflareAPIKey != "" && logflareSourceID != ""
}

func (p *LogflareProvider) GetInfoMessage() string {
	return "Default logging is configured for Logflare. To use other logging providers, submit a PR."
}

func (p *LogflareProvider) GenerateConfig(cfg *config.Config, imageName, envName string) (string, error) {
	// Read the Fluent-bit template
	templatePath := "deploy/configs/fluent-bit.conf"
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Fluent-bit template: %w", err)
	}

	// Render the template with configuration values
	rendered, err := templatex.RenderEnv(string(templateBytes), cfg)
	if err != nil {
		return "", fmt.Errorf("failed to render Fluent-bit template: %w", err)
	}

	return rendered, nil
}

package logging

import (
	"github.com/furiatona/azctl/internal/config"
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

	return generateConfigFromTemplate(templatePath, cfg, "Logflare")
}

package logging

import (
	"fmt"
	"os"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/templatex"
)

// DatadogProvider implements LoggingProvider for Datadog
// This is an example of how to add a new logging provider
type DatadogProvider struct{}

func (p *DatadogProvider) Name() string {
	return "Datadog"
}

func (p *DatadogProvider) IsEnabled(cfg *config.Config) bool {
	datadogAPIKey := cfg.Get("DATADOG_API_KEY")
	datadogSite := cfg.Get("DATADOG_SITE")
	return datadogAPIKey != "" && datadogSite != ""
}

func (p *DatadogProvider) GetInfoMessage() string {
	return "Datadog logging is enabled. Set DATADOG_API_KEY and DATADOG_SITE in Azure App Configuration."
}

func (p *DatadogProvider) GenerateConfig(cfg *config.Config, imageName, envName string) (string, error) {
	// This would read a different template for Datadog
	templatePath := "deploy/configs/fluent-bit-datadog.conf"
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Datadog Fluent-bit template: %w", err)
	}

	// Render the template with configuration values
	rendered, err := templatex.RenderEnv(string(templateBytes), cfg)
	if err != nil {
		return "", fmt.Errorf("failed to render Datadog Fluent-bit template: %w", err)
	}

	return rendered, nil
}

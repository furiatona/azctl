# Modular Logging System

This directory contains a modular logging system that allows easy integration with different logging platforms.

## Architecture

The logging system uses a provider-based architecture:

- `manager.go` - Main manager that handles different logging providers
- `logflare.go` - Logflare provider implementation
- `datadog.go` - Example Datadog provider implementation

## How to Add a New Logging Provider

1. **Create a new provider file** (e.g., `newrelic.go`):

```go
package logging

import (
    "azctl/internal/config"
    "azctl/internal/templatex"
    "fmt"
    "os"
)

type NewRelicProvider struct{}

func (p *NewRelicProvider) Name() string {
    return "New Relic"
}

func (p *NewRelicProvider) IsEnabled(cfg *config.Config) bool {
    apiKey := cfg.Get("NEWRELIC_API_KEY")
    accountID := cfg.Get("NEWRELIC_ACCOUNT_ID")
    return apiKey != "" && accountID != ""
}

func (p *NewRelicProvider) GetInfoMessage() string {
    return "New Relic logging is enabled. Set NEWRELIC_API_KEY and NEWRELIC_ACCOUNT_ID in Azure App Configuration."
}

func (p *NewRelicProvider) GenerateConfig(cfg *config.Config, imageName, envName string) (string, error) {
    templatePath := "deploy/configs/fluent-bit-newrelic.conf"
    templateBytes, err := os.ReadFile(templatePath)
    if err != nil {
        return "", fmt.Errorf("failed to read New Relic Fluent-bit template: %w", err)
    }
    
    rendered, err := templatex.RenderEnv(string(templateBytes), cfg)
    if err != nil {
        return "", fmt.Errorf("failed to render New Relic Fluent-bit template: %w", err)
    }
    
    return rendered, nil
}
```

2. **Register the provider** in `manager.go`:

```go
func NewManager() *Manager {
    return &Manager{
        providers: []LoggingProvider{
            &LogflareProvider{},
            &DatadogProvider{},
            &NewRelicProvider{}, // Add your new provider here
        },
    }
}
```

3. **Create the Fluent-bit template** (e.g., `deploy/configs/fluent-bit-newrelic.conf`):

```
[SERVICE]
    Flush        1
    Daemon       Off
    Log_Level    info

[INPUT]
    Name         tail
    Path         /var/log/app/{{ env "IMAGE_NAME" }}-{{ env "ENV_NAME" }}.log
    Tag          {{ env "IMAGE_NAME" }}
    Refresh_Interval 5
    Read_from_Head  true

[FILTER]
    Name         modify
    Match        {{ env "IMAGE_NAME" }}
    Add          service {{ env "IMAGE_NAME" }}
    Add          env {{ env "ENV_NAME" }}

[OUTPUT]
    Name         http
    Match        *
    Host         {{ env "NEWRELIC_ACCOUNT_ID" }}.http-inputs.newrelic.com
    Port         443
    URI          /v1/{{ env "NEWRELIC_API_KEY" }}
    Format       json
    TLS          On
```

## Provider Priority

Providers are checked in the order they are registered. The first enabled provider will be used.

## Configuration

Each provider requires specific environment variables to be set in Azure App Configuration:

- **Logflare**: `LOGFLARE_API_KEY`, `LOGFLARE_SOURCE_ID`
- **Datadog**: `DATADOG_API_KEY`, `DATADOG_SITE`
- **New Relic**: `NEWRELIC_API_KEY`, `NEWRELIC_ACCOUNT_ID`

## Contributing

To add a new logging provider:

1. Create the provider implementation
2. Add the provider to the manager
3. Create the Fluent-bit template
4. Update this README
5. Submit a PR

The system is designed to be easily extensible while maintaining backward compatibility.

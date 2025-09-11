package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

const envTrue = "true"

// Provider defines the interface for configuration providers
type Provider interface {
	Name() string
	Load(ctx context.Context) (map[string]string, error)
	Priority() int // Higher priority means higher precedence
}

// Config represents the application configuration
type Config struct {
	values map[string]string
	mu     sync.RWMutex
}

// New creates a new configuration instance
func New() *Config {
	return &Config{
		values: make(map[string]string),
	}
}

// Load loads configuration from multiple sources with proper precedence
func (c *Config) Load(ctx context.Context, envfile string, env string) error {
	providers := []Provider{
		&AzureAppConfigProvider{env: env},
		&EnvFileProvider{envfile: envfile},
		&EnvironmentProvider{},
	}

	// Sort providers by priority (highest first)
	for i := 0; i < len(providers)-1; i++ {
		for j := i + 1; j < len(providers); j++ {
			if providers[i].Priority() < providers[j].Priority() {
				providers[i], providers[j] = providers[j], providers[i]
			}
		}
	}

	// Load from each provider
	for _, provider := range providers {
		values, err := provider.Load(ctx)
		if err != nil {
			continue
		}

		c.mu.Lock()
		for k, v := range values {
			c.values[strings.ToUpper(k)] = v
		}
		c.mu.Unlock()
	}

	// Apply fallback logic for common variables
	c.applyFallbacks()

	return nil
}

// Get retrieves a configuration value by key
func (c *Config) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.values[strings.ToUpper(key)]; ok {
		return v
	}
	return ""
}

// Set sets a configuration value
func (c *Config) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[strings.ToUpper(key)] = value
}

// GetAll returns all configuration values
func (c *Config) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range c.values {
		result[k] = v
	}
	return result
}

// Require retrieves a required configuration value, panicking if not found
func (c *Config) Require(key string) string {
	v := c.Get(key)
	if v == "" {
		panic(fmt.Errorf("missing required configuration: %s", key))
	}
	return v
}

// Has checks if a configuration key exists
func (c *Config) Has(key string) bool {
	return c.Get(key) != ""
}

// applyFallbacks applies fallback logic for common variables
func (c *Config) applyFallbacks() {
	// Try to derive ACR_REGISTRY from other sources
	if c.Get("ACR_REGISTRY") == "" {
		if registry := c.Get("REGISTRY"); registry != "" {
			c.Set("ACR_REGISTRY", registry)
		}
	}

	// Map ACR_REGISTRY to IMAGE_REGISTRY for template compatibility
	if c.Get("IMAGE_REGISTRY") == "" {
		if acrRegistry := c.Get("ACR_REGISTRY"); acrRegistry != "" {
			c.Set("IMAGE_REGISTRY", acrRegistry)
		}
	}
}

// Validate validates that all required variables are present
func (c *Config) Validate(required []string) error {
	var missing []string
	for _, v := range required {
		if !c.Has(v) {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// EnvironmentProvider loads configuration from environment variables
type EnvironmentProvider struct{}

func (p *EnvironmentProvider) Name() string  { return "Environment" }
func (p *EnvironmentProvider) Priority() int { return 100 }

func (p *EnvironmentProvider) Load(ctx context.Context) (map[string]string, error) {
	values := make(map[string]string)
	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) == 2 {
			values[strings.ToUpper(kv[0])] = kv[1]
		}
	}
	return values, nil
}

// EnvFileProvider loads configuration from .env files
type EnvFileProvider struct {
	envfile string
}

func (p *EnvFileProvider) Name() string  { return "EnvFile" }
func (p *EnvFileProvider) Priority() int { return 50 }

func (p *EnvFileProvider) Load(ctx context.Context) (map[string]string, error) {
	if p.envfile == "" {
		return make(map[string]string), nil
	}

	// Skip .env loading in CI environments
	if os.Getenv("CI") == envTrue {
		return make(map[string]string), nil
	}

	if _, err := os.Stat(p.envfile); err != nil {
		return make(map[string]string), nil
	}

	values, err := godotenv.Read(p.envfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file %s: %w", p.envfile, err)
	}

	// Convert to uppercase keys
	result := make(map[string]string)
	for k, v := range values {
		result[strings.ToUpper(k)] = v
	}

	return result, nil
}

// AzureAppConfigProvider loads configuration from Azure App Configuration
type AzureAppConfigProvider struct {
	env string
}

func (p *AzureAppConfigProvider) Name() string  { return "AzureAppConfig" }
func (p *AzureAppConfigProvider) Priority() int { return 10 }

func (p *AzureAppConfigProvider) Load(ctx context.Context) (map[string]string, error) {
	// Skip if explicitly disabled
	if os.Getenv("APP_CONFIG_SKIP") == envTrue {
		return make(map[string]string), nil
	}

	name := os.Getenv("APP_CONFIG_NAME")
	if name == "" {
		name = os.Getenv("APP_CONFIG")
	}

	if name == "" {
		return make(map[string]string), nil
	}

	// Determine service name for Azure App Config
	serviceName := p.determineServiceName()

	// Get the label from APP_CONFIG_LABEL environment variable
	label := os.Getenv("APP_CONFIG_LABEL")
	if label == "" {
		// Fallback to environment name if no label is specified
		label = p.env
	}

	return fetchAzureAppConfigWithImage(ctx, name, label, serviceName)
}

// determineServiceName determines the service name for Azure App Config
func (p *AzureAppConfigProvider) determineServiceName() string {
	// In CI: auto-detect from GITHUB_REPOSITORY
	if os.Getenv("CI") == envTrue {
		if os.Getenv("GITHUB_ACTIONS") == envTrue {
			if repoName := os.Getenv("GITHUB_REPOSITORY"); repoName != "" {
				parts := strings.Split(repoName, "/")
				if len(parts) == 2 {
					return parts[1]
				}
			}
		}
	}

	// Use IMAGE_NAME from environment
	return os.Getenv("IMAGE_NAME")
}

// Global configuration instance
var globalConfig *Config
var configOnce sync.Once

// Init initializes the global configuration
func Init(ctx context.Context, envfile string, env string) error {
	var initErr error
	configOnce.Do(func() {
		globalConfig = New()
		initErr = globalConfig.Load(ctx, envfile, env)
	})
	return initErr
}

// Current returns the current global configuration
func Current() *Config {
	if globalConfig == nil {
		panic("configuration not initialized")
	}
	return globalConfig
}

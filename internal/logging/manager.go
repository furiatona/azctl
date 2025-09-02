package logging

import (
	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"
	"fmt"
	"os"
	"path/filepath"
)

// LoggingProvider defines the interface for different logging platforms
type LoggingProvider interface {
	Name() string
	GenerateConfig(cfg *config.Config, imageName, envName string) (string, error)
	IsEnabled(cfg *config.Config) bool
	GetInfoMessage() string
}

// Manager handles different logging providers
type Manager struct {
	providers []LoggingProvider
}

// NewManager creates a new logging manager with default providers
func NewManager() *Manager {
	return &Manager{
		providers: []LoggingProvider{
			&LogflareProvider{},
			// Add more providers here as they are implemented
		},
	}
}

// RegisterProvider adds a new logging provider
func (m *Manager) RegisterProvider(provider LoggingProvider) {
	m.providers = append(m.providers, provider)
}

// GenerateConfig generates configuration for the first enabled provider
func (m *Manager) GenerateConfig(cfg *config.Config, imageName, envName string) error {
	for _, provider := range m.providers {
		if provider.IsEnabled(cfg) {
			logx.Infof("Generating %s logging configuration...", provider.Name())
			logx.Infof(provider.GetInfoMessage())
			
			configContent, err := provider.GenerateConfig(cfg, imageName, envName)
			if err != nil {
				return fmt.Errorf("failed to generate %s config: %w", provider.Name(), err)
			}
			
			if err := writeConfigFile(configContent, imageName); err != nil {
				return fmt.Errorf("failed to write %s config: %w", provider.Name(), err)
			}
			
			return nil
		}
	}
	
	// No enabled providers found
	logx.Infof("No logging provider enabled. Available providers:")
	for _, provider := range m.providers {
		logx.Infof("  - %s: %s", provider.Name(), provider.GetInfoMessage())
	}
	return nil
}

// writeConfigFile writes the configuration to the appropriate location
func writeConfigFile(configContent, imageName string) error {
	// Create fluent-bit/etc directory if it doesn't exist
	configDir := "fluent-bit/etc"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create fluent-bit config directory: %w", err)
	}
	
	// Write the configuration file
	configPath := filepath.Join(configDir, fmt.Sprintf("%s.conf", imageName))
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write Fluent-bit config: %w", err)
	}
	
	logx.Infof("Fluent-bit configuration generated: %s", configPath)
	logx.Infof("This file will be mounted in the ACI container at /fluent-bit/etc/%s.conf", imageName)
	
	return nil
}

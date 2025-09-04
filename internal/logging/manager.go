package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/furiatona/azctl/internal/config"
	"github.com/furiatona/azctl/internal/logx"
	"github.com/furiatona/azctl/internal/runx"
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

			if err := writeConfigFile(configContent, imageName, cfg); err != nil {
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
func writeConfigFile(configContent, imageName string, cfg *config.Config) error {
	// Create fluent-bit/etc directory if it doesn't exist
	configDir := "fluent-bit/etc"
	if err := os.MkdirAll(configDir, 0755); err != nil { //nolint:gosec // acceptable permissions for config directory
		return fmt.Errorf("failed to create fluent-bit config directory: %w", err)
	}

	// Write the configuration file
	configPath := filepath.Join(configDir, fmt.Sprintf("%s.conf", imageName))
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write Fluent-bit config: %w", err)
	}

	logx.Infof("Fluent-bit configuration generated: %s", configPath)
	logx.Infof("This file will be mounted in the ACI container at /fluent-bit/etc/%s.conf", imageName)

	// Upload to Azure File Storage
	if err := uploadToAzureFileStorage(configPath, imageName, cfg); err != nil {
		return fmt.Errorf("failed to upload config to Azure File Storage: %w", err)
	}

	return nil
}

// uploadToAzureFileStorage uploads the configuration file to Azure File Storage
func uploadToAzureFileStorage(configPath, imageName string, cfg *config.Config) error {
	// Get required Azure Storage configuration
	storageAccount := cfg.Get("LOG_STORAGE_ACCOUNT")
	storageKey := cfg.Get("LOG_STORAGE_KEY")
	fluentbitConfigShare := cfg.Get("FLUENTBIT_CONFIG")

	if storageAccount == "" || storageKey == "" || fluentbitConfigShare == "" {
		logx.Warnf("Azure Storage configuration incomplete. Skipping upload.")
		logx.Warnf("Required: LOG_STORAGE_ACCOUNT, LOG_STORAGE_KEY, FLUENTBIT_CONFIG")
		return nil
	}

	logx.Infof("Uploading Fluent-bit configuration to Azure File Storage...")
	logx.Infof("Storage Account: %s", storageAccount)
	logx.Infof("File Share: %s", fluentbitConfigShare)
	logx.Infof("File: %s.conf", imageName)

	// Create file share if it doesn't exist
	ctx := context.Background()
	if err := createFileShareIfNotExists(ctx, storageAccount, storageKey, fluentbitConfigShare); err != nil {
		return fmt.Errorf("failed to create file share: %w", err)
	}

	// Upload file using Azure CLI
	args := []string{
		"storage", "file", "upload",
		"--account-name", storageAccount,
		"--account-key", storageKey,
		"--share-name", fluentbitConfigShare,
		"--source", configPath,
		"--path", fmt.Sprintf("%s.conf", imageName),
	}

	if err := runx.AZ(ctx, args...); err != nil {
		return fmt.Errorf("failed to upload file to Azure File Storage: %w", err)
	}

	logx.Infof("✅ Fluent-bit configuration uploaded successfully to Azure File Storage")
	logx.Infof("File will be available at: /fluent-bit/etc/%s.conf in the ACI container", imageName)

	return nil
}

// createFileShareIfNotExists creates the Azure File Storage share if it doesn't exist
func createFileShareIfNotExists(ctx context.Context, storageAccount, storageKey, shareName string) error {
	// Check if share exists
	checkArgs := []string{
		"storage", "share", "show",
		"--account-name", storageAccount,
		"--account-key", storageKey,
		"--name", shareName,
	}

	_, err := runx.AZOutput(ctx, checkArgs...)
	if err == nil {
		// Share exists, no need to create
		return nil
	}

	logx.Infof("Creating Azure File Storage share: %s", shareName)

	// Create the share
	createArgs := []string{
		"storage", "share", "create",
		"--account-name", storageAccount,
		"--account-key", storageKey,
		"--name", shareName,
		"--quota", "1", // 1 GB quota
	}

	if err := runx.AZ(ctx, createArgs...); err != nil {
		return fmt.Errorf("failed to create file share: %w", err)
	}

	logx.Infof("✅ Azure File Storage share created: %s", shareName)
	return nil
}

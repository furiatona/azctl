package logging

import (
	"os"
	"testing"

	"github.com/furiatona/azctl/internal/config"
)

// MockProvider is a test provider that doesn't require template files
type MockProvider struct{}

func (p *MockProvider) Name() string {
	return "Mock"
}

func (p *MockProvider) IsEnabled(cfg *config.Config) bool {
	return cfg.Get("MOCK_ENABLED") == "true"
}

func (p *MockProvider) GetInfoMessage() string {
	return "Mock provider for testing"
}

func (p *MockProvider) GenerateConfig(cfg *config.Config, imageName, envName string) (string, error) {
	// Return a simple mock config
	return `[SERVICE]
    Flush        1
    Daemon       Off
    Log_Level    info

[INPUT]
    Name         tail
    Path         /var/log/app/test.log
    Tag          test

[OUTPUT]
    Name         stdout
    Match        *`, nil
}

func TestManager_GenerateConfig_NoProviderEnabled(t *testing.T) {
	// Create a config with no logging providers enabled
	cfg := config.New()

	manager := NewManager()

	// This should not return an error, just log that no providers are enabled
	err := manager.GenerateConfig(cfg, "test-app", "dev")
	if err != nil {
		t.Errorf("Expected no error when no providers enabled, got: %v", err)
	}
}

func TestManager_GenerateConfig_MissingStorageConfig(t *testing.T) {
	// Create a config with mock provider enabled but missing storage config
	cfg := config.New()
	cfg.Set("MOCK_ENABLED", "true")
	// Missing LOG_STORAGE_ACCOUNT, LOG_STORAGE_KEY, FLUENTBIT_CONFIG

	manager := NewManager()
	manager.RegisterProvider(&MockProvider{})

	// This should not return an error, just log a warning and skip upload
	err := manager.GenerateConfig(cfg, "test-app", "dev")
	if err != nil {
		t.Errorf("Expected no error when storage config missing, got: %v", err)
	}

	// Check if local file was still created
	if _, err := os.Stat("fluent-bit/etc/test-app.conf"); os.IsNotExist(err) {
		t.Error("Expected config file to be created locally even without storage config")
	}

	// Clean up
	if err := os.RemoveAll("fluent-bit"); err != nil {
		t.Logf("Failed to clean up fluent-bit directory: %v", err)
	}
}

func TestManager_GenerateConfig_WithStorageConfig(t *testing.T) {
	// Create a config with mock provider and storage config
	cfg := config.New()
	cfg.Set("MOCK_ENABLED", "true")
	cfg.Set("LOG_STORAGE_ACCOUNT", "test-storage")
	cfg.Set("LOG_STORAGE_KEY", "test-key")
	cfg.Set("FLUENTBIT_CONFIG", "test-config")

	manager := NewManager()
	manager.RegisterProvider(&MockProvider{})

	// This should not return an error, even if upload fails due to missing Azure CLI or invalid credentials
	err := manager.GenerateConfig(cfg, "test-app", "dev")
	if err != nil {
		// If the error is about Azure CLI or credentials, that's expected in test environment
		expectedError := "failed to write Mock config: failed to upload config to Azure File Storage: " +
			"failed to create file share: failed to create file share: az command failed: exit status 1"
		if err.Error() == expectedError {
			t.Logf("Expected Azure CLI error in test environment: %v", err)
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Check if local file was created
	if _, err := os.Stat("fluent-bit/etc/test-app.conf"); os.IsNotExist(err) {
		t.Error("Expected config file to be created locally")
	}

	// Clean up
	if err := os.RemoveAll("fluent-bit"); err != nil {
		t.Logf("Failed to clean up fluent-bit directory: %v", err)
	}
}

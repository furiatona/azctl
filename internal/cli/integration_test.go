package cli

import (
	"context"
	"os"
	"testing"
)

func TestACRCommandValidation(t *testing.T) {
	// Clean environment
	defer func() {
		for _, v := range []string{"REGISTRY", "ACR_RESOURCE_GROUP", "IMAGE_NAME", "IMAGE_TAG"} {
			// nolint:errcheck // os.Unsetenv rarely fails in test cleanup
			os.Unsetenv(v)
		}
	}()

	// Test missing variables
	err := Execute(context.Background(), []string{"acr"})
	if err == nil {
		t.Error("expected error for missing required variables")
	}
	// Check that the error contains the expected message about missing required variables
	if err.Error() != "missing required variable: IMAGE_NAME" &&
		err.Error() != "missing required variable: IMAGE_TAG" &&
		err.Error() != "missing required variable: REGISTRY" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCIEnvironmentDetection(t *testing.T) {
	// Test CI environment detection
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expected: true,
		},
		{
			name: "Azure Pipeline",
			envVars: map[string]string{
				"AZURE_PIPELINE": "true",
			},
			expected: true,
		},
		{
			name: "GitLab CI",
			envVars: map[string]string{
				"GITLAB_CI": "true",
			},
			expected: true,
		},
		{
			name: "Generic CI",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: true,
		},
		{
			name:     "Not CI",
			envVars:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			result := isCIEnvironment()
			if result != tt.expected {
				t.Errorf("isCIEnvironment() = %v, want %v", result, tt.expected)
			}

			// Clean up
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestEnvironmentDetectionFromCI(t *testing.T) {
	// Test environment detection from CI context
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "GitHub Actions - dev branch",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
				"GITHUB_REF":     "refs/heads/dev",
			},
			expected: "dev",
		},
		{
			name: "GitHub Actions - staging branch",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
				"GITHUB_REF":     "refs/heads/staging",
			},
			expected: "staging",
		},
		{
			name: "GitHub Actions - main branch",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
				"GITHUB_REF":     "refs/heads/main",
			},
			expected: "prod",
		},
		{
			name: "Azure Pipeline",
			envVars: map[string]string{
				"AZURE_PIPELINE":     "true",
				"SYSTEM_ENVIRONMENT": "staging",
			},
			expected: "staging",
		},
		{
			name: "GitLab CI",
			envVars: map[string]string{
				"GITLAB_CI":           "true",
				"CI_ENVIRONMENT_NAME": "production",
			},
			expected: "production",
		},
		{
			name: "Explicit environment",
			envVars: map[string]string{
				"ENVIRONMENT": "dev",
			},
			expected: "dev",
		},
		{
			name:     "No environment",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			result := detectEnvironmentFromCI()
			if result != tt.expected {
				t.Errorf("detectEnvironmentFromCI() = %v, want %v", result, tt.expected)
			}

			// Clean up
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestACIDeployCommandValidation(t *testing.T) {
	// Clean environment
	defer func() {
		for _, v := range []string{"AZURE_RESOURCE_GROUP", "CONTAINER_GROUP_NAME", "IMAGE_NAME", "IMAGE_TAG"} {
			// nolint:errcheck // os.Unsetenv rarely fails in test cleanup
			os.Unsetenv(v)
		}
	}()

	// Test missing variables (will fail on template not found first, but that's expected)
	err := Execute(context.Background(), []string{"aci"})
	if err == nil {
		t.Error("expected error for missing template or variables")
	}
	// Should fail on template not found since we don't have deploy/manifests/aci.json in test
	if err.Error() != "template not found: deploy/manifests/aci.json" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestACIDryRun(t *testing.T) {
	// Set minimal required environment variables
	t.Setenv("AZURE_RESOURCE_GROUP", "test-rg")
	t.Setenv("CONTAINER_GROUP_NAME", "test-container")
	t.Setenv("LOCATION", "eastus")
	t.Setenv("OS_TYPE", "Linux")
	t.Setenv("ACI_PORT", "8080")
	t.Setenv("ACI_CPU", "1")
	t.Setenv("ACI_MEMORY", "2")
	t.Setenv("IMAGE_REGISTRY", "testregistry")
	t.Setenv("IMAGE_NAME", "testapp")
	t.Setenv("IMAGE_TAG", "v1.0.0")
	t.Setenv("ACR_USERNAME", "testuser")
	t.Setenv("ACR_PASSWORD", "testpass")
	t.Setenv("DNS_NAME_LABEL", "test-app")

	// Application-specific variables (from template)
	t.Setenv("FIREBASE_KEY", "test-key")
	t.Setenv("FIREBASE_URL", "https://test.firebase.co")
	t.Setenv("SAGEMAKER_OPENAI_MODEL", "test-model")
	t.Setenv("SAGEMAKER_OPENAI_API_KEY", "test-api-key")
	t.Setenv("OPENAI_SAGEMAKER_EMBEDDINGS_ENDPOINT", "https://test.example.com")
	t.Setenv("LOG_SHARE_NAME", "test-logs")
	t.Setenv("LOG_STORAGE_ACCOUNT", "teststorage")
	t.Setenv("LOG_STORAGE_KEY", "test-storage-key")
	t.Setenv("FLUENTBIT_CONFIG_SHARE", "test-config")

	// Test dry-run (should succeed and create file)
	err := Execute(context.Background(), []string{"aci", "--dry-run", "--template", "../../deploy/manifests/aci.json"})
	if err != nil {
		t.Fatalf("dry-run should not error: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(".azctl/aci-dry-run.json"); os.IsNotExist(err) {
		t.Error("dry-run should create .azctl/aci-dry-run.json")
	}

	// Clean up
	if err := os.RemoveAll(".azctl"); err != nil {
		t.Logf("failed to clean up .azctl directory: %v", err)
	}
}

func TestHelpCommands(t *testing.T) {
	tests := [][]string{
		{"--help"},
		{"acr", "--help"},
		{"aci", "--help"},
		{"webapp", "--help"},
	}

	for _, args := range tests {
		err := Execute(context.Background(), args)
		// Help commands should not return an error
		if err != nil {
			t.Errorf("help command %v should not error: %v", args, err)
		}
	}
}

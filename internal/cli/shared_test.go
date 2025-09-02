package cli

import (
	"os"
	"testing"
)

func TestImageNameDetectionFromCI(t *testing.T) {
	// Test GitHub Actions
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")

	detected := detectImageNameFromCI()
	if detected != "test-repo" {
		t.Errorf("Expected 'test-repo', got '%s'", detected)
	}

	// Clean up
	// nolint:errcheck // os.Unsetenv rarely fails in test cleanup
	os.Unsetenv("GITHUB_ACTIONS")
	// nolint:errcheck // os.Unsetenv rarely fails in test cleanup
	os.Unsetenv("GITHUB_REPOSITORY")
}

func TestImageTagDetectionFromCI(t *testing.T) {
	// Test GitHub Actions
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_SHA", "abc123def456")

	detected := detectImageTagFromCI()
	if detected != "abc123def456" {
		t.Errorf("Expected 'abc123def456', got '%s'", detected)
	}

	// Clean up
	// nolint:errcheck // os.Unsetenv rarely fails in test cleanup
	os.Unsetenv("GITHUB_ACTIONS")
	// nolint:errcheck // os.Unsetenv rarely fails in test cleanup
	os.Unsetenv("GITHUB_SHA")
}

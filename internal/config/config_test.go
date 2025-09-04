package config

import (
	"context"
	"os"
	"sync"
	"testing"
)

func TestConfigPrecedence(t *testing.T) {
	// Clean up environment
	defer func() {
		//nolint:errcheck // os.Unsetenv rarely fails in test cleanup
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("CI") //nolint:errcheck
	}()

	// Test 1: Environment variable only
	//nolint:errcheck // os.Setenv rarely fails in test setup
	os.Setenv("TEST_VAR", "from_env") //nolint:gosec // acceptable in test setup
	err := Init(context.Background(), "", "")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cfg := Current()
	if got := cfg.Get("TEST_VAR"); got != "from_env" {
		t.Errorf("expected 'from_env', got %q", got)
	}

	// Test 2: Missing variable
	if got := cfg.Get("NONEXISTENT"); got != "" {
		t.Errorf("expected empty string for missing var, got %q", got)
	}
}

func TestConfigRequire(t *testing.T) {
	// Reset global config for this test
	globalConfig = nil
	configOnce = sync.Once{}

	defer func() {
		//nolint:errcheck // os.Unsetenv rarely fails in test cleanup
		os.Unsetenv("REQUIRED_VAR")
		//nolint:errcheck // os.Unsetenv rarely fails in test cleanup
		os.Unsetenv("MISSING_REQUIRED")
		// Reset global config after test
		globalConfig = nil
		configOnce = sync.Once{}
	}()

	//nolint:errcheck // os.Setenv rarely fails in test setup
	os.Setenv("REQUIRED_VAR", "value") //nolint:gosec // acceptable in test setup
	_ = Init(context.Background(), "", "")
	cfg := Current()

	// Should not panic
	if got := cfg.Require("REQUIRED_VAR"); got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}

	// Should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing required var")
		}
	}()
	cfg.Require("MISSING_REQUIRED")
}

func TestCISkipsEnvFile(t *testing.T) {
	defer func() {
		//nolint:errcheck // os.Unsetenv rarely fails in test cleanup
		os.Unsetenv("CI")
		os.Unsetenv("TEST_CI_VAR") //nolint:errcheck
	}()

	// Create a temp .env file
	envFile := t.TempDir() + "/.env"
	if err := os.WriteFile(envFile, []byte("TEST_CI_VAR=from_dotenv\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with CI=true - should skip .env
	//nolint:errcheck // os.Setenv rarely fails in test setup
	os.Setenv("CI", "true")              //nolint:gosec // acceptable in test setup
	os.Setenv("TEST_CI_VAR", "from_env") //nolint:errcheck // acceptable in test setup

	err := Init(context.Background(), envFile, "")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cfg := Current()
	if got := cfg.Get("TEST_CI_VAR"); got != "from_env" {
		t.Errorf("in CI mode, expected env var to win, got %q", got)
	}
}

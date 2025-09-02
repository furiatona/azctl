package validation

import (
	"context"
	"strings"
	"testing"

	"github.com/furiatona/azctl/internal/config"
)

func TestRequiredVars(t *testing.T) {
	// Initialize empty config
	_ = config.Init(context.Background(), "", "")
	cfg := config.Current()

	// Test with missing variables
	err := RequiredVars(cfg, []string{"MISSING_VAR1", "MISSING_VAR2"})
	if err == nil {
		t.Error("expected error for missing variables")
	}
	if !strings.Contains(err.Error(), "MISSING_VAR1") || !strings.Contains(err.Error(), "MISSING_VAR2") {
		t.Errorf("error should mention missing variables: %v", err)
	}

	// Test with all variables present
	cfg.Set("VAR1", "value1")
	cfg.Set("VAR2", "value2")
	err = RequiredVars(cfg, []string{"VAR1", "VAR2"})
	if err != nil {
		t.Errorf("expected no error when all vars present, got: %v", err)
	}

	// Test with some missing, some present
	err = RequiredVars(cfg, []string{"VAR1", "MISSING_VAR"})
	if err == nil {
		t.Error("expected error when some vars missing")
	}
	if !strings.Contains(err.Error(), "MISSING_VAR") {
		t.Errorf("error should mention missing variable: %v", err)
	}
	if strings.Contains(err.Error(), "VAR1") {
		t.Errorf("error should not mention present variable: %v", err)
	}
}

func TestACRRequiredVars(t *testing.T) {
	vars := ACRRequiredVars()
	expected := []string{"ACR_REGISTRY", "ACR_RESOURCE_GROUP", "IMAGE_NAME", "IMAGE_TAG"}

	if len(vars) != len(expected) {
		t.Errorf("expected %d vars, got %d", len(expected), len(vars))
	}

	for _, exp := range expected {
		found := false
		for _, v := range vars {
			if v == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing expected variable: %s", exp)
		}
	}
}

func TestWebAppRequiredVars(t *testing.T) {
	vars := WebAppRequiredVars()
	required := []string{"RESOURCE_GROUP", "ACR_REGISTRY", "IMAGE_NAME", "IMAGE_TAG"}

	for _, req := range required {
		found := false
		for _, v := range vars {
			if v == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("WebApp required vars missing: %s", req)
		}
	}
}

func TestACIRequiredVars(t *testing.T) {
	vars := ACIRequiredVars()
	required := []string{"RESOURCE_GROUP", "CONTAINER_GROUP_NAME", "IMAGE_NAME", "IMAGE_TAG"}

	for _, req := range required {
		found := false
		for _, v := range vars {
			if v == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ACI required vars missing: %s", req)
		}
	}
}

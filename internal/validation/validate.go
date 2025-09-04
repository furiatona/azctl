package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/furiatona/azctl/internal/config"
)

// Validator defines the interface for validation rules
type Validator interface {
	Validate(cfg *config.Config) error
	Name() string
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Name     string
	Required []string
	Patterns map[string]string
	Custom   func(cfg *config.Config) error
}

// ValidationEngine handles validation of configuration
type ValidationEngine struct {
	rules []ValidationRule
}

// NewEngine creates a new validation engine
func NewEngine() *ValidationEngine {
	return &ValidationEngine{
		rules: []ValidationRule{},
	}
}

// AddRule adds a validation rule to the engine
func (e *ValidationEngine) AddRule(rule ValidationRule) {
	e.rules = append(e.rules, rule)
}

// Validate validates configuration against all rules
func (e *ValidationEngine) Validate(cfg *config.Config) error {
	var errors []string

	for _, rule := range e.rules {
		if err := e.validateRule(cfg, rule); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", rule.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateRule validates a single rule
func (e *ValidationEngine) validateRule(cfg *config.Config, rule ValidationRule) error {
	// Check required fields
	for _, field := range rule.Required {
		if !cfg.Has(field) {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Check patterns
	for field, pattern := range rule.Patterns {
		if value := cfg.Get(field); value != "" {
			matched, err := regexp.MatchString(pattern, value)
			if err != nil {
				return fmt.Errorf("invalid regex pattern for %s: %w", field, err)
			}
			if !matched {
				return fmt.Errorf("field %s does not match pattern %s", field, pattern)
			}
		}
	}

	// Run custom validation
	if rule.Custom != nil {
		if err := rule.Custom(cfg); err != nil {
			return err
		}
	}

	return nil
}

// Predefined validation rules
var (
	// ACRValidation validates Azure Container Registry configuration
	ACRValidation = ValidationRule{
		Name: "ACR Configuration",
		Required: []string{
			"ACR_REGISTRY",
			"ACR_RESOURCE_GROUP",
			"IMAGE_NAME",
			"IMAGE_TAG",
		},
		Patterns: map[string]string{
			"ACR_REGISTRY": `^[a-zA-Z0-9]+$`,
			"IMAGE_NAME":   `^[a-zA-Z0-9_-]+$`,
			"IMAGE_TAG":    `^[a-zA-Z0-9._-]+$`,
		},
		Custom: func(cfg *config.Config) error {
			// Validate ACR registry format
			registry := cfg.Get("ACR_REGISTRY")
			if !strings.HasSuffix(registry, ".azurecr.io") {
				return fmt.Errorf("ACR_REGISTRY should end with .azurecr.io")
			}
			return nil
		},
	}

	// WebAppValidation validates Azure Web App configuration
	WebAppValidation = ValidationRule{
		Name: "WebApp Configuration",
		Required: []string{
			"RESOURCE_GROUP",
			"ACR_REGISTRY",
			"IMAGE_NAME",
			"IMAGE_TAG",
		},
		Patterns: map[string]string{
			"RESOURCE_GROUP": `^[a-zA-Z0-9_-]+$`,
			"WEBAPP_NAME":    `^[a-zA-Z0-9_-]+$`,
		},
	}

	// ACIValidation validates Azure Container Instance configuration
	ACIValidation = ValidationRule{
		Name: "ACI Configuration",
		Required: []string{
			"RESOURCE_GROUP",
			"CONTAINER_GROUP_NAME",
			"LOCATION",
			"OS_TYPE",
			"DNS_NAME_LABEL",
			"ACI_PORT",
			"ACI_CPU",
			"ACI_MEMORY",
			"ACR_REGISTRY",
			"IMAGE_NAME",
			"IMAGE_TAG",
			"ACR_USERNAME",
			"ACR_PASSWORD",
		},
		Patterns: map[string]string{
			"RESOURCE_GROUP":       `^[a-zA-Z0-9_-]+$`,
			"CONTAINER_GROUP_NAME": `^[a-zA-Z0-9_-]+$`,
			"DNS_NAME_LABEL":       `^[a-zA-Z0-9-]+$`,
			"ACI_PORT":             `^\d+$`,
			"ACI_CPU":              `^\d+(\.\d+)?$`,
			"ACI_MEMORY":           `^\d+(\.\d+)?$`,
			"OS_TYPE":              `^(Linux|Windows)$`,
		},
		Custom: func(cfg *config.Config) error {
			// Validate CPU and memory values
			cpu := cfg.Get("ACI_CPU")
			memory := cfg.Get("ACI_MEMORY")

			if cpu != "" {
				if cpuFloat, err := parseFloat(cpu); err != nil {
					return fmt.Errorf("invalid ACI_CPU value: %s", cpu)
				} else if cpuFloat <= 0 || cpuFloat > 4 {
					return fmt.Errorf("ACI_CPU must be between 0.1 and 4.0")
				}
			}

			if memory != "" {
				if memoryFloat, err := parseFloat(memory); err != nil {
					return fmt.Errorf("invalid ACI_MEMORY value: %s", memory)
				} else if memoryFloat <= 0 || memoryFloat > 16 {
					return fmt.Errorf("ACI_MEMORY must be between 0.1 and 16.0")
				}
			}

			return nil
		},
	}

	// SecurityValidation validates security-related configuration
	SecurityValidation = ValidationRule{
		Name: "Security Configuration",
		Custom: func(cfg *config.Config) error {
			// Check for sensitive data in plain text
			sensitiveFields := []string{
				"ACR_PASSWORD",
				"AZURE_OPENAI_API_KEY",
				"FIREBASE_KEY",
				"SUPABASE_KEY",
			}

			for _, field := range sensitiveFields {
				if value := cfg.Get(field); value != "" {
					if len(value) < 8 {
						return fmt.Errorf("sensitive field %s appears to be too short", field)
					}
				}
			}

			return nil
		},
	}
)

// parseFloat safely parses a string to float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, fmt.Errorf("failed to parse float: %w", err)
}

// Convenience functions for backward compatibility

// RequiredVars validates that all required variables are present in config
func RequiredVars(cfg *config.Config, vars []string) error {
	var missing []string
	for _, v := range vars {
		if !cfg.Has(v) {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ACRRequiredVars returns the list of variables required for ACR operations
func ACRRequiredVars() []string {
	return []string{
		"ACR_REGISTRY",
		"ACR_RESOURCE_GROUP",
		"IMAGE_NAME",
		"IMAGE_TAG",
	}
}

// WebAppRequiredVars returns the list of variables required for WebApp deployment
func WebAppRequiredVars() []string {
	return []string{
		"RESOURCE_GROUP",
		"ACR_REGISTRY",
		"IMAGE_NAME",
		"IMAGE_TAG",
	}
}

// ACIRequiredVars returns the list of variables required for ACI deployment
func ACIRequiredVars() []string {
	return []string{
		"RESOURCE_GROUP",
		"CONTAINER_GROUP_NAME",
		"LOCATION",
		"OS_TYPE",
		"DNS_NAME_LABEL",
		"ACI_PORT",
		"ACI_CPU",
		"ACI_MEMORY",
		"ACR_REGISTRY",
		"IMAGE_NAME",
		"IMAGE_TAG",
		"ACR_USERNAME",
		"ACR_PASSWORD",
	}
}

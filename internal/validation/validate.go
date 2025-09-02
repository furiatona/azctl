package validation

import (
	"fmt"
	"strings"

	"github.com/furiatona/azctl/internal/config"
)

// RequiredVars validates that all required variables are present in config.
// Returns a user-friendly error message listing all missing variables.
func RequiredVars(cfg *config.Config, vars []string) error {
	var missing []string
	for _, v := range vars {
		if cfg.Get(v) == "" {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ACRRequiredVars returns the list of variables required for ACR operations.
func ACRRequiredVars() []string {
	return []string{
		"ACR_REGISTRY",
		"ACR_RESOURCE_GROUP",
		"IMAGE_NAME",
		"IMAGE_TAG",
	}
}

// WebAppRequiredVars returns the list of variables required for WebApp deployment.
func WebAppRequiredVars() []string {
	return []string{
		"RESOURCE_GROUP",
		"ACR_REGISTRY",
		"IMAGE_NAME",
		"IMAGE_TAG",
	}
}

// ACIRequiredVars returns the list of variables required for ACI deployment.
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

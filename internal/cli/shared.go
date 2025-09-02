package cli

import (
	"os"
	"strings"
)

// isCIEnvironment detects if we're running in a CI environment
func isCIEnvironment() bool {
	// Check for common CI environment variables
	ciVars := []string{"CI", "GITHUB_ACTIONS", "AZURE_PIPELINE", "GITLAB_CI", "JENKINS_URL", "TRAVIS", "CIRCLECI"}
	for _, envVar := range ciVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}

// detectEnvironmentFromCI detects the current environment from CI context
func detectEnvironmentFromCI() string {
	// Try to detect from GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		ref := os.Getenv("GITHUB_REF")
		if strings.HasPrefix(ref, "refs/heads/") {
			branch := strings.TrimPrefix(ref, "refs/heads/")
			switch branch {
			case "dev", "development":
				return "dev"
			case "staging":
				return "staging"
			case "main", "master", "prod", "production":
				return "prod"
			}
		}
	}

	// Try to detect from Azure Pipeline
	if os.Getenv("AZURE_PIPELINE") == "true" {
		// Azure Pipeline environment variables
		if env := os.Getenv("SYSTEM_ENVIRONMENT"); env != "" {
			return strings.ToLower(env)
		}
	}

	// Try to detect from GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		if env := os.Getenv("CI_ENVIRONMENT_NAME"); env != "" {
			return strings.ToLower(env)
		}
	}

	// Try to detect from explicit environment variable
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return strings.ToLower(env)
	}

	return ""
}

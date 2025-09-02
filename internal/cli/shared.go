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

// detectImageNameFromCI detects the image name from CI context
func detectImageNameFromCI() string {
	// Try to detect from GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		if repoName := os.Getenv("GITHUB_REPOSITORY"); repoName != "" {
			// Extract repository name from "owner/repo" format
			parts := strings.Split(repoName, "/")
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	// Try to detect from Azure Pipeline
	if os.Getenv("AZURE_PIPELINE") == "true" {
		if buildRepoName := os.Getenv("BUILD_REPOSITORY_NAME"); buildRepoName != "" {
			return buildRepoName
		}
	}

	// Try to detect from GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		if projectName := os.Getenv("CI_PROJECT_NAME"); projectName != "" {
			return projectName
		}
	}

	return ""
}

// detectImageTagFromCI detects the image tag from CI context
func detectImageTagFromCI() string {
	// Try to detect from GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		if sha := os.Getenv("GITHUB_SHA"); sha != "" {
			return sha
		}
	}

	// Try to detect from Azure Pipeline
	if os.Getenv("AZURE_PIPELINE") == "true" {
		if buildId := os.Getenv("BUILD_BUILDID"); buildId != "" {
			return buildId
		}
		if sourceVersion := os.Getenv("BUILD_SOURCEVERSION"); sourceVersion != "" {
			return sourceVersion
		}
	}

	// Try to detect from GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		if commitSha := os.Getenv("CI_COMMIT_SHA"); commitSha != "" {
			return commitSha
		}
	}

	return ""
}

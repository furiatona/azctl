package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/furiatona/azctl/internal/logx"

	"github.com/joho/godotenv"
)

// Provider precedence (low -> high):
// 1) Azure App Config (future extension hook)
// 2) .env file
// 3) Environment Variables
// 4) CLI Flags (handled in commands before querying Config)

type Config struct {
	values map[string]string
}

var current *Config

func Init(ctx context.Context, envfile string, env string) error {
	// Merge order (low -> high): Azure App Config -> .env -> OS env
	merged := map[string]string{}

	// First pass: Load .env to get IMAGE_NAME for Azure App Config
	envImageName := ""
	if os.Getenv("CI") != "true" && envfile != "" {
		if _, err := os.Stat(envfile); err == nil {
			logx.Infof("[DEBUG] Pre-loading .env file for IMAGE_NAME: %s", envfile)
			if m, err := godotenv.Read(envfile); err != nil {
				logx.Warnf("failed reading .env: %v", err)
			} else {
				if imageName, ok := m["IMAGE_NAME"]; ok {
					envImageName = imageName
					logx.Infof("[DEBUG] Found IMAGE_NAME in .env: '%s'", envImageName)
				}
			}
		}
	}

	// Azure App Configuration (lowest precedence) - now with environment and IMAGE_NAME from .env
	if os.Getenv("APP_CONFIG_SKIP") != "true" {
		name := os.Getenv("APP_CONFIG_NAME")
		if name == "" {
			name = os.Getenv("APP_CONFIG") // Fallback to APP_CONFIG
		}
		label := env // Use environment as label for Azure App Config
		// Use IMAGE_NAME from .env if available, otherwise from OS env
		imageName := envImageName
		if imageName == "" {
			imageName = os.Getenv("IMAGE_NAME")
		}
		logx.Infof("[DEBUG] Using IMAGE_NAME for Azure App Config: '%s', environment: '%s', app config: '%s'", imageName, env, name)
		if name != "" {
			if kvs, err := fetchAzureAppConfigWithImage(ctx, name, label, imageName); err != nil {
				logx.Warnf("azure appconfig fetch failed: %v", err)
			} else {
				logx.Infof("[DEBUG] Loaded %d variables from Azure App Config", len(kvs))
				for k, v := range kvs {
					logx.Infof("[DEBUG] From Azure App Config: %s='%s'", k, v)
					merged[strings.ToUpper(k)] = v
				}
			}
		} else {
			logx.Infof("[DEBUG] No APP_CONFIG_NAME or APP_CONFIG set, skipping Azure App Configuration")
		}
	}

	// .env (middle precedence) - load only when not in CI
	if os.Getenv("CI") != "true" {
		if envfile != "" {
			if _, err := os.Stat(envfile); err == nil {
				logx.Infof("[DEBUG] Loading .env file: %s", envfile)
				if m, err := godotenv.Read(envfile); err != nil {
					logx.Warnf("failed reading .env: %v", err)
				} else {
					logx.Infof("[DEBUG] Loaded %d variables from .env file", len(m))
					for k, v := range m {
						logx.Infof("[DEBUG] From .env: %s='%s'", strings.ToUpper(k), v)
						merged[strings.ToUpper(k)] = v
					}
				}
			} else {
				logx.Infof("[DEBUG] .env file not found: %s", envfile)
			}
		} else {
			logx.Infof("[DEBUG] No envfile specified, skipping .env loading")
		}
	} else {
		logx.Infof("[DEBUG] CI=true, skipping .env file loading")
	}

	// OS environment variables (highest among config sources)
	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) == 2 {
			merged[strings.ToUpper(kv[0])] = kv[1]
		}
	}

	// Apply fallback logic for common variables
	if merged["ACR_REGISTRY"] == "" {
		// Try to derive ACR_REGISTRY from other sources
		if registry := merged["REGISTRY"]; registry != "" {
			merged["ACR_REGISTRY"] = registry
			logx.Infof("[DEBUG] Derived ACR_REGISTRY from REGISTRY: '%s'", registry)
		}
	}

	current = &Config{values: merged}
	return nil
}

func Current() *Config {
	if current == nil {
		panic("config not initialized")
	}
	return current
}

func (c *Config) Get(key string) string {
	if v, ok := c.values[strings.ToUpper(key)]; ok {
		return v
	}
	return ""
}

func (c *Config) Set(key, value string) {
	if c.values == nil {
		c.values = make(map[string]string)
	}
	c.values[strings.ToUpper(key)] = value
}

func (c *Config) GetAll() map[string]string {
	if c.values == nil {
		return map[string]string{}
	}
	return c.values
}

func (c *Config) Require(key string) string {
	v := c.Get(key)
	if v == "" {
		panic(fmt.Errorf("missing required variable: %s", key))
	}
	return v
}

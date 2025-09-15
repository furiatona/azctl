# azctl Setup Guide

A comprehensive guide for setting up and configuring `azctl` for Azure container and web app deployment workflows.

## Table of Contents

- [Quick Start](#quick-start)
- [Simple Setup](#simple-setup)
- [Environment Configuration](#environment-configuration)
- [Workflow Examples](#workflow-examples)
- [Advanced Setup](#advanced-setup)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)

## Quick Start

The fastest way to get started with `azctl` in your GitHub Actions workflow:

```yaml
- uses: azure/login@v2
  with:
    creds: ${{ secrets.AZURE_CREDENTIALS }}

- uses: furiatona/azctl@v1
```

## Simple Setup

### Prerequisites

- Azure subscription with appropriate permissions
- GitHub repository with Actions enabled
- Azure Container Registry (ACR)
- Azure App Configuration (optional but recommended)

### Required Environment Variables

#### For Azure Container Instance (ACI)
```yaml
APP_CONFIG: ${{ vars.APP_CONFIG }}                    # Azure App Configuration name
APP_CONFIG_GLOBAL_KEY: ${{ vars.APP_CONFIG_GLOBAL_KEY }}  # Global config key (e.g., "global-configurations")
```

#### For Azure Web App
```yaml
WEBAPP_NAME: ${{ vars.WEBAPP_NAME }}                  # Your web app name
APP_SERVICE_PLAN: ${{ vars.APP_SERVICE_PLAN }}        # Your App Service plan
APP_CONFIG: ${{ vars.APP_CONFIG }}                    # Azure App Configuration name
APP_CONFIG_GLOBAL_KEY: ${{ vars.APP_CONFIG_GLOBAL_KEY }}  # Global config key
```

### GitHub Repository Setup

1. **Go to Repository Settings → Environments**
2. **Create environments** (e.g., `dev`, `staging`, `production`)
3. **Add required variables** for each environment:

   **Example for dev environment:**
   ```
   APP_CONFIG=yourorg-app-conf-dev
   APP_CONFIG_GLOBAL_KEY=global-configurations
   WEBAPP_NAME=mywebapp-dev
   APP_SERVICE_PLAN=ASP-<resourcegroup>-bffc
   ```

### Azure App Configuration Setup

1. **Create Azure App Configuration instance**
2. **Configure global settings** under the key specified in `APP_CONFIG_GLOBAL_KEY`:

   ```json
   {
     "RESOURCE_GROUP": "your-deployment-rg",
     "ACR_REGISTRY": "your-registry.azurecr.io",
     "LOCATION": "eastus",
     "ACR_RESOURCE_GROUP": "your-acr-rg"
   }
   ```

3. **Configure service-specific settings** using your `IMAGE_NAME` as the key:

   ```json
   {
     "FIREBASE_URL": "https://your-project.firebase.co",
     "FIREBASE_KEY": "your-firebase-key",
     "AZURE_OPENAI_MODEL": "gpt-4",
     "AZURE_OPENAI_API_KEY": "your-openai-key",
     "OPENAI_AZURE_EMBEDDINGS_ENDPOINT": "https://your-endpoint.openai.azure.com",
     "NEXT_PUBLIC_URL": "example.com"
   }
   ```

   > **Note:** Variables with `NEXT_PUBLIC_*` prefix will be injected into the container images.

## Environment Configuration

For basic usage, the environment variables configured in GitHub Actions (as shown in Simple Setup) are sufficient. The azctl action will automatically handle environment detection and configuration loading.

If you need to run azctl locally or require custom configurations, see the [Advanced Setup](#advanced-setup) section for detailed environment variable configuration.

## Workflow Examples

### Azure Container Instance Deployment

```yaml
name: Deploy to Azure Container Instance

on:
  push:
    branches: [dev, staging]

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ github.ref_name }}
    env:
      APP_CONFIG: ${{ vars.APP_CONFIG }}
      APP_CONFIG_GLOBAL_KEY: ${{ vars.APP_CONFIG_GLOBAL_KEY }}
    steps:
      - uses: actions/checkout@v4

      - uses: azure/login@v2
        with:
          creds: ${{ secrets.AZURE_CREDENTIALS }}

      - uses: furiatona/azctl@v1

      - name: Build & Push to ACR
        run: azctl acr --env ${{ github.ref_name }}

      - name: Deploy to Azure Container Instance
        run: azctl aci --env ${{ github.ref_name }}
```

### Azure Web App Deployment

```yaml
name: Deploy to Azure Web App

on:
  push:
    branches: [dev, staging]

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ github.ref_name }}
    env:
      WEBAPP_NAME: ${{ vars.WEBAPP_NAME }}
      APP_SERVICE_PLAN: ${{ vars.APP_SERVICE_PLAN }}
      APP_CONFIG: ${{ vars.APP_CONFIG }}
      APP_CONFIG_GLOBAL_KEY: ${{ vars.APP_CONFIG_GLOBAL_KEY }}
    steps:
      - uses: actions/checkout@v4

      - uses: azure/login@v2
        with:
          creds: ${{ secrets.AZURE_CREDENTIALS }}

      - uses: furiatona/azctl@v1

      - name: Build & Push to ACR
        run: azctl acr --env ${{ github.ref_name }}

      - name: Deploy to Azure Web App
        run: azctl webapp --env ${{ github.ref_name }}
```

## Advanced Setup

### Local Development Setup

For local development and testing:

1. **Copy environment files:**
   ```bash
   cp env.dev.example .env.dev
   cp env.staging.example .env.staging
   cp env.prod.example .env.prod
   ```

2. **Edit with your values:**
   ```bash
   nano .env.dev
   ```

### Core Environment Variables

#### Required Variables
```bash
# Azure Container Registry
ACR_REGISTRY=your-registry.azurecr.io
ACR_RESOURCE_GROUP=your-acr-rg

# Image Configuration
IMAGE_NAME=your-app-name
IMAGE_TAG=latest

# Azure Resources
RESOURCE_GROUP=your-deployment-rg
LOCATION=eastus

# Azure App Configuration
APP_CONFIG=your-app-config-name
APP_CONFIG_RESOURCE_GROUP=your-app-config-rg
```

#### ACI-Specific Variables
```bash
# Container Instance Configuration
CONTAINER_GROUP_NAME=your-aci-group
ACI_CPU=1
ACI_MEMORY=2
ACI_PORT=8080
OS_TYPE=Linux

# Application Variables
FIREBASE_URL=https://your-project.firebase.co
FIREBASE_KEY=your-firebase-key
AZURE_OPENAI_MODEL=gpt-4
AZURE_OPENAI_API_KEY=your-openai-key
OPENAI_AZURE_EMBEDDINGS_ENDPOINT=https://your-endpoint.openai.azure.com

# Logging Configuration
LOG_STORAGE_NAME=your-log-storage
LOG_STORAGE_ACCOUNT=your-storage-account
LOG_STORAGE_KEY=your-storage-key
FLUENTBIT_CONFIG=your-fluentbit-config
```

#### WebApp-Specific Variables
```bash
# Web App Configuration
WEBAPP_NAME=your-webapp-name
APP_SERVICE_PLAN=your-service-plan
DNS_NAME_LABEL=your-dns-label
```

### Azure App Configuration with Environment Labels

For multi-environment setups using a single App Configuration instance:

```bash
# Add environment-specific labels
az appconfig kv set --name your-app-config-name \
  --key global-configurations \
  --label staging \
  --value '{"RESOURCE_GROUP": "rg-staging", "ACR_REGISTRY": "staging.azurecr.io"}'

az appconfig kv set --name your-app-config-name \
  --key global-configurations \
  --label prod \
  --value '{"RESOURCE_GROUP": "rg-prod", "ACR_REGISTRY": "prod.azurecr.io"}'
```

### Custom Templates

Create custom deployment templates for specific requirements:

```bash
# Use custom template
azctl aci --template ./custom-aci.json --env staging
```

### Multi-Environment Configuration

For complex setups with separate App Configuration instances:

```bash
# Environment-specific App Config instances
APP_CONFIG_DEV=your-dev-config
APP_CONFIG_STAGING=your-staging-config
APP_CONFIG_PROD=your-prod-config
```

### Development Workflow

```bash
# 1. Build and push image
azctl acr --env dev

# 2. Deploy to ACI for testing
azctl aci --env dev --dry-run  # Test configuration
azctl aci --env dev            # Deploy

# 3. Deploy to WebApp for staging
azctl webapp --env staging
```

## Troubleshooting

### Common Issues

#### Missing Environment Variables
```bash
# Check required variables
azctl aci --env staging --dry-run --verbose
```

#### Azure App Configuration Access
```bash
# Verify access
az appconfig kv list --name your-app-config-name
```

#### Template Rendering Issues
```bash
# Debug template variables
azctl aci --env staging --dry-run --verbose
```

### Debug Tools

Use the provided debug script for Azure App Configuration issues:

```bash
# Debug App Configuration fetching
./debug_appconfig.sh
```

## Security Best Practices

1. **Use Azure Key Vault** for sensitive values
2. **Enable RBAC** on Azure resources
3. **Use managed identities** when possible
4. **Rotate credentials** regularly
5. **Audit access** to App Configuration
6. **Use environment-specific secrets** in GitHub Actions

## Support

For issues and questions:
- Check the [Environment Configuration Guide](ENVIRONMENT_CONFIG.md)
- Review the [troubleshooting section](#troubleshooting)
- Open an issue on GitHub

---

## Additional Resources

- [Azure Container Instances Documentation](https://docs.microsoft.com/en-us/azure/container-instances/)
- [Azure Web Apps Documentation](https://docs.microsoft.com/en-us/azure/app-service/)
- [Azure App Configuration Documentation](https://docs.microsoft.com/en-us/azure/azure-app-configuration/)
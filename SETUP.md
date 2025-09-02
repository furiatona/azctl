# Setup Guide

This guide provides detailed instructions for setting up and configuring `azctl` for your Azure container deployment workflows.

## Prerequisites

- Azure CLI installed and authenticated
- Go 1.21+ (for development)
- Docker (for building images)
- Azure Container Registry (ACR)
- Azure App Configuration (optional, for centralized config)

## Initial Setup

### 1. Environment Configuration

Create environment-specific configuration files:

```bash
# Copy example files
cp env.dev.example .env.dev
cp env.staging.example .env.staging  
cp env.prod.example .env.prod

# Edit with your values
nano .env.dev
```

### 2. Required Environment Variables

#### Core Variables (Required)
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

# Azure App Configuration (optional)
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

# Application Variables (customize per your app)
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

### 3. Azure App Configuration Setup

For centralized configuration management:

#### Create App Configuration Instance
```bash
# Create resource group
az group create --name your-app-config-rg --location eastus

# Create App Configuration
az appconfig create --name your-app-config-name \
  --resource-group your-app-config-rg \
  --location eastus \
  --sku Standard
```

#### Configure Keys

**Global Configuration Key:**
```json
{
  "RESOURCE_GROUP": "your-deployment-rg",
  "ACR_REGISTRY": "your-registry.azurecr.io",
  "LOCATION": "eastus",
  "ACR_RESOURCE_GROUP": "your-acr-rg"
}
```

**Service-Specific Key (use your IMAGE_NAME):**
```json
{
  "FIREBASE_URL": "https://your-project.firebase.co",
  "FIREBASE_KEY": "your-firebase-key",
  "AZURE_OPENAI_MODEL": "gpt-4",
  "AZURE_OPENAI_API_KEY": "your-openai-key",
  "OPENAI_AZURE_EMBEDDINGS_ENDPOINT": "https://your-endpoint.openai.azure.com"
}
```

#### Optional: Environment Labels

If you want to use a single App Configuration instance for multiple environments:

```bash
# Add labels to keys for environment separation
az appconfig kv set --name your-app-config-name \
  --key global-configurations \
  --label staging \
  --value '{"RESOURCE_GROUP": "rg-staging", "ACR_REGISTRY": "staging.azurecr.io"}'

az appconfig kv set --name your-app-config-name \
  --key global-configurations \
  --label prod \
  --value '{"RESOURCE_GROUP": "rg-prod", "ACR_REGISTRY": "prod.azurecr.io"}'
```

### 4. Azure Container Registry Setup

```bash
# Create ACR
az acr create --name your-registry \
  --resource-group your-acr-rg \
  --sku Basic \
  --admin-enabled true

# Get credentials
az acr credential show --name your-registry
```

### 5. Template Customization

#### ACI Template (`deploy/manifests/aci.json`)

Customize the template for your application:

```json
{
  "name": "{{ env \"CONTAINER_GROUP_NAME\" }}",
  "type": "Microsoft.ContainerInstance/containerGroups",
  "apiVersion": "2021-10-01",
  "location": "{{ env \"LOCATION\" }}",
  "properties": {
    "containers": [
      {
        "name": "{{ env \"IMAGE_NAME\" }}",
        "properties": {
          "image": "{{ env \"IMAGE_REGISTRY\" }}/{{ env \"IMAGE_NAME\" }}:{{ env \"IMAGE_TAG\" }}",
          "resources": {
            "requests": {
              "cpu": "{{ env \"CPU\" }}",
              "memoryInGB": "{{ env \"MEMORY\" }}"
            }
          },
          "ports": [
            {
              "port": "{{ env \"PORT\" }}"
            }
          ],
                     "environmentVariables": [
             {
               "name": "FIREBASE_URL",
               "value": "{{ env \"FIREBASE_URL\" }}"
             },
             {
               "name": "FIREBASE_KEY", 
               "value": "{{ env \"FIREBASE_KEY\" }}"
             }
           ]
        }
      }
    ],
    "osType": "{{ env \"OS_TYPE\" }}",
    "restartPolicy": "Always"
  }
}
```

#### Fluent-bit Configuration (`deploy/configs/fluent-bit.conf`)

Customize logging configuration:

```ini
[INPUT]
    Name tail
    Path /var/log/containers/{{ env "IMAGE_NAME" }}*.log
    Parser docker
    Tag kube.*
    Mem_Buf_Limit 5MB
    Skip_Long_Lines On

[OUTPUT]
    Name azure
    Match *
    Customer_ID your-workspace-id
    Shared_Key your-workspace-key
    Log_Type your-log-type
```

## Usage Examples

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

### CI/CD Integration

```bash
# In CI pipeline - environment auto-detected
azctl acr                    # Build and push
azctl aci --resource-group $RESOURCE_GROUP  # Deploy
```

### Environment-Specific Deployments

```bash
# Development
azctl aci --env dev --resource-group rg-dev

# Staging  
azctl aci --env staging --resource-group rg-staging

# Production
azctl aci --env prod --resource-group rg-prod
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

### Debug Scripts

Use the provided debug script to troubleshoot Azure App Configuration:

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

## Advanced Configuration

### Custom Templates

Create custom deployment templates:

```bash
# Use custom template
azctl aci --template ./custom-aci.json --env staging
```

### Multi-Environment Setup

For complex multi-environment setups:

```bash
# Environment-specific App Config instances
APP_CONFIG_DEV=your-dev-config
APP_CONFIG_STAGING=your-staging-config  
APP_CONFIG_PROD=your-prod-config
```

### Integration with CI/CD

Example GitHub Actions workflow:

```yaml
name: Deploy to Azure
on:
  push:
    branches: [main, staging]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build and Deploy
        run: |
          make build
          ./bin/azctl acr
          ./bin/azctl aci --resource-group ${{ secrets.RESOURCE_GROUP }}
        env:
          ACR_REGISTRY: ${{ secrets.ACR_REGISTRY }}
          RESOURCE_GROUP: ${{ secrets.RESOURCE_GROUP }}
```

## Support

For issues and questions:
- Check the [Environment Configuration Guide](ENVIRONMENT_CONFIG.md)
- Review [troubleshooting section](#troubleshooting)
- Open an issue on GitHub

# azctl Setup Guide

This guide provides detailed setup instructions for azctl, a Go CLI tool for Azure container deployment workflows.

## Prerequisites

- Azure CLI installed and authenticated
- Go 1.22+ (for building from source)
- Docker (for building container images)

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the latest binary from the releases:

```bash
# Linux AMD64
curl -L https://github.com/furiatona/azctl/releases/latest/download/azctl_linux_amd64 -o azctl
chmod +x azctl
sudo mv azctl /usr/local/bin/

# macOS AMD64
curl -L https://github.com/furiatona/azctl/releases/latest/download/azctl_darwin_amd64 -o azctl
chmod +x azctl
sudo mv azctl /usr/local/bin/

# macOS ARM64
curl -L https://github.com/furiatona/azctl/releases/latest/download/azctl_darwin_arm64 -o azctl
chmod +x azctl
sudo mv azctl /usr/local/bin/

# Windows AMD64
# Download azctl_windows_amd64.exe from https://github.com/furiatona/azctl/releases/latest/download/
```

### Option 2: Build from Source

```bash
git clone https://github.com/furiatona/azctl
cd azctl
make build
./bin/azctl --help
```

## Configuration

### Precedence Order (highest to lowest)

1. **CLI Flags** - Explicit command-line arguments
2. **Environment Variables** - Shell environment and CI variables  
3. **.env File** - Local development only (skipped when `CI=true`)
4. **Azure App Configuration** - Centralized defaults (optional)

### Environment Variable Naming

#### Local Development (`.env` file)
Use environment-specific prefixed variables for clarity:

```bash
# Development environment
DEV_RESOURCE_GROUP=rg-myapp-dev
DEV_APP_CONFIG=myapp-app-conf-dev
DEV_WEBAPP_NAME=myapp-website-dev

# Staging environment  
STAGING_RESOURCE_GROUP=rg-myapp-staging
STAGING_APP_CONFIG=myapp-app-conf
STAGING_WEBAPP_NAME=myapp-website-staging

# Production environment
PROD_RESOURCE_GROUP=rg-myapp-prod
PROD_APP_CONFIG=myapp-app-conf-prod
PROD_WEBAPP_NAME=myapp-website-prod
```

#### CI Environments (GitHub Actions, Azure Pipelines)
Use non-prefixed variables for cleaner configuration:

```yaml
# GitHub Actions workflow
env:
  RESOURCE_GROUP: rg-myapp-staging
  APP_CONFIG: myapp-app-conf
  WEBAPP_NAME: myapp-website-staging
  ACI_SUPABASE_KEY: ${{ secrets.SUPABASE_KEY }}
  ACI_FIREBASE_KEY: ${{ secrets.FIREBASE_KEY }}
```

The tool automatically:
1. **Detects CI environment** (GitHub Actions, Azure Pipeline, GitLab CI, etc.)
2. **Auto-detects environment** (dev/staging/prod) from branch name
3. **Maps prefixed variables** to non-prefixed versions
4. **Uses appropriate configuration** for the detected environment

### Required Variables

#### ACR Commands
- `REGISTRY` - ACR registry name
- `ACR_RESOURCE_GROUP` - Resource group containing ACR
- `IMAGE_NAME` - Container image name  
- `IMAGE_TAG` - Container image tag

#### WebApp Deployment
- `AZURE_RESOURCE_GROUP` - Target resource group
- `REGISTRY` - ACR registry name
- `IMAGE_NAME` - Container image name
- `IMAGE_TAG` - Container image tag

#### ACI Deployment
- `AZURE_RESOURCE_GROUP` - Target resource group
- `CONTAINER_GROUP_NAME` - ACI container group name
- `IMAGE_REGISTRY` - ACR registry name
- `IMAGE_NAME` - Container image name
- `IMAGE_TAG` - Container image tag
- `ACR_USERNAME` - Registry username
- `ACR_PASSWORD` - Registry password

## Configuration Setup

### 1. Create Environment File

```bash
cp env.sample .env
```

### 2. Configure Azure Resources

Edit `.env` with your Azure configuration:

```bash
# Azure Configuration
REGISTRY=myregistry
AZURE_RESOURCE_GROUP=my-resource-group
ACR_RESOURCE_GROUP=my-acr-resource-group
LOCATION=eastus

# Container Configuration
IMAGE_NAME=my-app
IMAGE_TAG=latest
CONTAINER_GROUP_NAME=my-app-container
DNS_NAME_LABEL=my-app-dev

# ACI Configuration
OS_TYPE=Linux
ACI_PORT=8080
ACI_CPU=1
ACI_MEMORY=2

# Registry Credentials
ACR_USERNAME=myregistry
ACR_PASSWORD=your-acr-password

# Application Environment Variables
FIREBASE_KEY=your-FIREBASE-key
FIREBASE_URL=https://your-project.FIREBASE.co/
SAGEMAKER_OPENAI_MODEL=text-embedding-3-small
SAGEMAKER_OPENAI_API_KEY=your-openai-key
OPENAI_SAGEMAKER_EMBEDDINGS_ENDPOINT=https://example.com

# Logging Configuration (optional)
LOG_SHARE_NAME=logshare-dev
LOG_STORAGE_ACCOUNT=mystorageaccount
LOG_STORAGE_KEY=your-storage-key
FLUENTBIT_CONFIG_SHARE=fluentbit-config-dev

# WebApp Configuration (optional)
WEBAPP_NAME=my-app-webapp
DEV_WEBAPP_NAME=my-app-dev
STAGING_WEBAPP_NAME=my-app-staging
PROD_WEBAPP_NAME=my-app-prod
APP_SERVICE_PLAN=my-app-service-plan
DEV_APP_SERVICE_PLAN=my-app-dev-plan
STAGING_APP_SERVICE_PLAN=my-app-staging-plan
PROD_APP_SERVICE_PLAN=my-app-prod-plan

# Azure App Configuration (optional)
APP_CONFIG_NAME=my-app-config
APP_CONFIG_LABEL=dev
APP_CONFIG_SKIP=false
APP_CONFIG_DEBUG=false
```

### 3. Azure Resource Setup

#### Azure Container Registry (ACR)

```bash
# Create ACR
az acr create --resource-group my-acr-resource-group --name myregistry --sku Basic

# Get ACR credentials
az acr credential show --name myregistry
```

#### Azure Storage Account (for logging)

```bash
# Create storage account
az storage account create --name mystorageaccount --resource-group my-resource-group --location eastus --sku Standard_LRS

# Create file shares
az storage share create --name logshare-dev --account-name mystorageaccount
az storage share create --name fluentbit-config-dev --account-name mystorageaccount

# Get storage key
az storage account keys list --resource-group my-resource-group --account-name mystorageaccount
```

## Usage Examples

### Build and Push to ACR

```bash
# Basic usage
azctl acr

# With custom parameters
azctl acr --registry myregistry --image myapp --tag v1.0.0 --resource-group my-rg

# Environment-specific deployment
azctl acr --env dev
azctl acr --env staging
azctl acr --env production
```

### Deploy to Azure Container Instances (ACI)

```bash
# Basic deployment
azctl aci --resource-group my-rg

# With custom template
azctl aci --template ./my-aci-template.json --env production

# Dry run (generate JSON without deploying)
azctl aci --dry-run --env staging --resource-group staging-rg
```

### Deploy to Azure Web Apps

```bash
# Deploy to existing WebApp
azctl webapp --env staging --resource-group my-rg

# Deploy with custom WebApp name
azctl webapp --env production --name my-custom-webapp
```

## Template System

ACI deployments use Go templates with environment variable substitution. The default template is located at `deploy/manifests/aci.json`.

### Custom Template Example

```json
{
  "name": "{{ env \"CONTAINER_GROUP_NAME\" }}",
  "location": "{{ env \"LOCATION\" }}",
  "properties": {
    "containers": [{
      "name": "app",
      "properties": {
        "image": "{{ env \"IMAGE_REGISTRY\" }}.azurecr.io/{{ env \"IMAGE_NAME\" }}:{{ env \"IMAGE_TAG\" }}",
        "environmentVariables": [
          { "name": "API_KEY", "value": "{{ env \"API_KEY\" }}" }
        ]
      }
    }]
  }
}
```

## Environment-Based Configuration

All commands support environment-specific configuration via the `--env` flag:

```bash
# Load dev environment config
azctl acr --env dev
azctl aci --env dev  
azctl webapp --env dev

# Load staging environment config
azctl acr --env staging
azctl aci --env staging
azctl webapp --env staging

# Load production environment config
azctl acr --env production
azctl aci --env production
azctl webapp --env production
```

When `--env` is specified, azctl will:

1. **Load Azure App Configuration**: Fetch from `app-config-{env}` with label `{env}`
2. **Load global-configurations**: JSON object with global settings
3. **Load image-specific keys**: Keys containing the `IMAGE_NAME` value
4. **Apply precedence**: CLI flags > Environment variables > .env > Azure App Configuration

## CI/CD Integration

Set environment variables in your CI system. The `.env` file is automatically skipped when `CI=true`.

### GitHub Actions Example

```yaml
name: Deploy
on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download azctl
        run: |
          curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl-linux-amd64 -o azctl
          chmod +x azctl
      
      - name: Deploy to ACR
        env:
          REGISTRY: ${{ secrets.REGISTRY }}
          IMAGE_NAME: my-app
          IMAGE_TAG: ${{ github.sha }}
          ACR_RESOURCE_GROUP: ${{ secrets.ACR_RESOURCE_GROUP }}
        run: ./azctl acr
      
      - name: Deploy to ACI
        env:
          AZURE_RESOURCE_GROUP: ${{ secrets.AZURE_RESOURCE_GROUP }}
          CONTAINER_GROUP_NAME: my-app-container
          IMAGE_REGISTRY: ${{ secrets.REGISTRY }}
          IMAGE_NAME: my-app
          IMAGE_TAG: ${{ github.sha }}
          ACR_USERNAME: ${{ secrets.ACR_USERNAME }}
          ACR_PASSWORD: ${{ secrets.ACR_PASSWORD }}
        run: ./azctl aci
```

## CI/CD Integration

Set environment variables in your CI system. The `.env` file is automatically skipped when `CI=true`.

### GitHub Actions Example

```yaml
name: Deploy to Azure

on:
  push:
    branches: [staging, main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: azure/login@v2
        with:
          creds: ${{ secrets.AZURE_CREDENTIALS }}
      
      - name: Download azctl
        run: |
          curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl_linux_amd64 -o azctl
          chmod +x azctl
      
      - name: Deploy to ACI
        env:
          RESOURCE_GROUP: rg-myapp-${{ github.ref_name }}
          APP_CONFIG: myapp-app-conf-${{ github.ref_name == 'main' && 'prod' || github.ref_name }}
          ACI_SUPABASE_KEY: ${{ secrets.SUPABASE_KEY }}
        run: |
          ./azctl aci  # Environment auto-detected!
```

### Azure Pipeline Example

```yaml
trigger:
  branches:
    include:
    - staging
    - main

pool:
  vmImage: 'ubuntu-latest'

variables:
  RESOURCE_GROUP: 'rg-myapp-$(Build.SourceBranchName)'
  APP_CONFIG: 'myapp-app-conf-$(Build.SourceBranchName)'

steps:
- task: AzureCLI@2
  inputs:
    azureSubscription: 'MyAzureSubscription'
    scriptLocation: 'inlineScript'
    inlineScript: |
      # Download azctl
      curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl_linux_amd64 -o azctl
      chmod +x azctl
      
      # Deploy (environment auto-detected)
      ./azctl aci
```

### Basic CI/CD Example

```yaml
name: Deploy
on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download azctl
        run: |
          curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl_linux_amd64 -o azctl
          chmod +x azctl
      
      - name: Deploy to ACR
        env:
          REGISTRY: ${{ secrets.REGISTRY }}
          IMAGE_NAME: my-app
          IMAGE_TAG: ${{ github.sha }}
          ACR_RESOURCE_GROUP: ${{ secrets.ACR_RESOURCE_GROUP }}
        run: ./azctl acr
      
      - name: Deploy to ACI
        env:
          AZURE_RESOURCE_GROUP: ${{ secrets.AZURE_RESOURCE_GROUP }}
          CONTAINER_GROUP_NAME: my-app-container
          IMAGE_REGISTRY: ${{ secrets.REGISTRY }}
          IMAGE_NAME: my-app
          IMAGE_TAG: ${{ github.sha }}
          ACR_USERNAME: ${{ secrets.ACR_USERNAME }}
          ACR_PASSWORD: ${{ secrets.ACR_PASSWORD }}
        run: ./azctl aci
```

## Troubleshooting

### Common Issues

1. **Missing Environment Variables**
   ```bash
   Error: missing required environment variables: REGISTRY, IMAGE_NAME
   ```
   Solution: Set the required environment variables in your `.env` file or as environment variables.

2. **ACI Deployment Validation Failed**
   ```bash
   Error: ACI deployment validation failed: missing required environment variables: CONTAINER_GROUP_NAME, ACR_PASSWORD
   ```
   Solution: Ensure all required ACI variables are set.

3. **Azure Authentication Issues**
   ```bash
   Error: az: command not found
   ```
   Solution: Install and authenticate Azure CLI.

### Debug Mode

Use `--verbose` flag for detailed output:

```bash
azctl aci --verbose --dry-run
```

### Dry Run

Use `--dry-run` to inspect the generated configuration without deploying:

```bash
azctl aci --dry-run --env staging
# Outputs to: .azctl/aci-dry-run.json (git-ignored)
```

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

### Linting

```bash
make lint
```

### Cross-platform Release Build

```bash
make release
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality  
4. Run `make test lint` 
5. Submit a pull request

## License

MIT License - see [LICENSE](../LICENSE) file for details.

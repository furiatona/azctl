# azctl

A production-ready Go CLI tool that wraps Azure CLI commands for container deployment workflows.

## Features

- **ACR Integration**: Build and push images to Azure Container Registry
- **WebApp Deployment**: Deploy containers to Azure Web Apps with automatic creation/update
- **ACI Deployment**: Deploy containers with sidecar support using JSON templates  
- **Config Management**: Multi-source configuration with proper precedence
- **Environment Support**: Seamless local development and CI/CD integration
- **CI Environment Detection**: Automatic environment detection and variable mapping
- **Validation**: Clear error messages for missing configuration

## Installation

### From Release (Recommended)

Download the latest binary from [dl.furiatona.dev](https://dl.furiatona.dev/azctl/):

```bash
# Linux AMD64
curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl_linux_amd64 -o azctl
chmod +x azctl
sudo mv azctl /usr/local/bin/

# macOS AMD64
curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl_darwin_amd64 -o azctl
chmod +x azctl
sudo mv azctl /usr/local/bin/

# macOS ARM64
curl -L https://dl.furiatona.dev/azctl/v0.2.0/azctl_darwin_arm64 -o azctl
chmod +x azctl
sudo mv azctl /usr/local/bin/

# Windows AMD64
# Download azctl_windows_amd64.exe from https://dl.furiatona.dev/azctl/v0.2.0/
```

### From Source

```bash
git clone https://github.com/furiatona/azctl
cd azctl
make build
./bin/azctl --help
```

## Quick Start

1. **Setup Configuration**:
   ```bash
   cp env.sample .env
   # Edit .env with your Azure configuration
   ```

2. **Build and Push to ACR**:
   ```bash
   azctl acr
   ```

3. **Deploy to ACI**:
   ```bash
   azctl aci
   ```

4. **Deploy to WebApp**:
   ```bash
   azctl webapp
   ```

> **📖 Need more detailed setup instructions?** See [SETUP.md](SETUP.md) for comprehensive configuration and deployment guides.

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
DEV_RESOURCE_GROUP=rg-swarm-dev
DEV_APP_CONFIG=chiswarm-app-conf-dev
DEV_WEBAPP_NAME=swarm-website-dev

# Staging environment  
STAGING_RESOURCE_GROUP=rg-swarm-staging
STAGING_APP_CONFIG=chiswarm-app-conf
STAGING_WEBAPP_NAME=swarm-website-staging

# Production environment
PROD_RESOURCE_GROUP=rg-swarm-prod
PROD_APP_CONFIG=chiswarm-app-conf-prod
PROD_WEBAPP_NAME=swarm-website-prod
```

#### CI Environments (GitHub Actions, Azure Pipelines)
Use non-prefixed variables for cleaner configuration:

```yaml
# GitHub Actions workflow
env:
  RESOURCE_GROUP: rg-swarm-staging
  APP_CONFIG: chiswarm-app-conf
  WEBAPP_NAME: swarm-website-staging
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

## Usage Examples

### Build and Push to ACR

```bash
# Using environment variables or .env
azctl acr

# Using CLI flags
azctl acr --registry myregistry --image myapp --tag v1.0.0

# With environment-specific config
azctl acr --env dev
azctl acr --env staging
azctl acr --env production

# In CI - environment auto-detected from branch name
azctl acr  # Auto-detects 'staging' from 'staging' branch
```

### Deploy to WebApp

```bash
# Deploy to existing WebApp
azctl webapp --env staging --resource-group my-rg

# Deploy with custom WebApp name
azctl webapp --env production --name my-custom-webapp

# In CI - environment auto-detected
azctl webapp --resource-group my-rg  # Auto-detects environment
```

### Deploy to ACI

```bash
# Deploy with default template
azctl aci --resource-group my-rg

# Deploy with custom template  
azctl aci --template ./my-aci-template.json --env production

# Dry run - generate JSON without deploying
azctl aci --dry-run --env staging --resource-group staging-rg

# In CI - environment auto-detected
azctl aci --resource-group my-rg  # Auto-detects environment
```

## CI/CD Integration

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
          RESOURCE_GROUP: rg-swarm-${{ github.ref_name }}
          APP_CONFIG: chiswarm-app-conf-${{ github.ref_name == 'main' && 'prod' || github.ref_name }}
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
  RESOURCE_GROUP: 'rg-swarm-$(Build.SourceBranchName)'
  APP_CONFIG: 'chiswarm-app-conf-$(Build.SourceBranchName)'

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

## Development

```bash
# Run tests
make test

# Build binary
make build

# Lint code
make lint

# Cross-platform release build
make release
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality  
4. Run `make test lint` 
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

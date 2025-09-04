# Professional Azure Container Deployment CLI Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/furiatona/azctl)](https://goreportcard.com/report/github.com/furiatona/azctl)
[![Go Version](https://img.shields.io/github/go-mod/go-version/furiatona/azctl)](https://golang.org)
[![License](https://img.shields.io/github/license/furiatona/azctl)](LICENSE)
[![Release](https://img.shields.io/github/v/release/furiatona/azctl)](https://github.com/furiatona/azctl/releases/latest)
[![GitHub Marketplace](https://img.shields.io/badge/GitHub-Marketplace-blue.svg)](https://github.com/marketplace/actions/azctl-installer)

**azctl** is a production-ready Go CLI tool that provides a seamless interface for Azure container deployment workflows. Built with enterprise-grade practices, it offers robust configuration management, comprehensive validation, and seamless CI/CD integration.

## ✨ Features

### 🚀 **Core Capabilities**
- **ACR Integration**: Build and push images to Azure Container Registry with intelligent caching
- **WebApp Deployment**: Deploy containers to Azure Web Apps with automatic creation/update
- **ACI Deployment**: Deploy containers with sidecar support using JSON templates
- **Multi-Environment Support**: Seamless local development and CI/CD integration

### 🔧 **Advanced Features**
- **Intelligent Configuration Management**: Multi-source configuration with proper precedence
- **CI Environment Detection**: Automatic environment detection and variable mapping
- **Comprehensive Validation**: Clear error messages for missing configuration
- **Professional Logging**: Structured logging with multiple output formats
- **Security First**: Secure handling of sensitive data and credentials

### 🏗️ **Enterprise Ready**
- **Zero-Downtime Deployments**: Smart deployment strategies for different environments
- **Rollback Capabilities**: Built-in rollback mechanisms for failed deployments
- **Health Checks**: Automated health monitoring and validation
- **Audit Trail**: Comprehensive logging for compliance and debugging

## 📦 Installation

### From GitHub Release (Recommended)

Download the latest binary from [GitHub Releases](https://github.com/furiatona/azctl/releases):

```bash
# Visit the latest release page and download the appropriate binary:
# https://github.com/furiatona/azctl/releases/latest
#
# Example filenames:
# - azctl_v2.0.0_linux_amd64
# - azctl_v2.0.0_darwin_amd64  
# - azctl_v2.0.0_darwin_arm64
# - azctl_v2.0.0_windows_amd64.exe

# After downloading, make executable and move to PATH:
chmod +x azctl_v2.0.0_darwin_arm64  # adjust filename as needed
sudo mv azctl_v2.0.0_darwin_arm64 /usr/local/bin/azctl
```

### Using Package Managers

#### Homebrew (macOS/Linux)
```bash
brew install furiatona/tap/azctl
```

#### Scoop (Windows)
```powershell
scoop bucket add furiatona https://github.com/furiatona/scoop-bucket
scoop install azctl
```

### In GitHub Workflows

Use azctl directly in your GitHub Actions workflows:

```yaml
- name: Use azctl
  uses: furiatona/azctl@v2

- name: Build and push to ACR
  run: azctl acr --env production
```

### From Source

```bash
git clone https://github.com/furiatona/azctl
cd azctl
make build
./bin/azctl --help
```

## 🚀 Quick Start

### 1. **Setup Configuration**
```bash
# Copy example environment file
cp env.dev.example .env.dev
# Edit .env.dev with your Azure configuration
```

### 2. **Build and Push to ACR**
```bash
azctl acr --env dev
```

### 3. **Deploy to ACI**
```bash
azctl aci --env dev
```

### 4. **Deploy to WebApp**
```bash
azctl webapp --env dev
```

> **📖 Need detailed setup instructions?** See [SETUP.md](SETUP.md) for comprehensive configuration and deployment guides.

## 📚 Usage Examples

### Build and Push to ACR

```bash
# Using environment-specific .env files
azctl acr --env dev
azctl acr --env staging
azctl acr --env prod

# Using CLI flags
azctl acr --registry myregistry --image myapp --tag v1.0.0

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

## 🛠️ Development

### Prerequisites
- Go 1.22+
- Docker (for building images)
- Azure CLI (for Azure operations)

### Development Commands

```bash
# Run tests with coverage
make test

# Build binary
make build

# Lint code
make lint

# Cross-platform release build
make release

# Run integration tests
make test-integration

# Generate documentation
make docs
```

### Project Structure

```
azctl/
├── cmd/azctl/          # Main application entry point
├── internal/           # Internal packages
│   ├── cli/           # CLI command implementations
│   ├── config/        # Configuration management
│   ├── validation/    # Input validation
│   ├── logging/       # Logging infrastructure
│   └── runx/          # External command execution
├── deploy/            # Deployment templates and configs
├── docs/             # Documentation
└── scripts/          # Build and deployment scripts
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Run `make test lint`
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Standards

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write comprehensive tests (aim for >90% coverage)
- Update documentation for new features
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Getting Help

- 📖 **Documentation**: [SETUP.md](SETUP.md), [ENVIRONMENT_CONFIG.md](ENVIRONMENT_CONFIG.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/furiatona/azctl/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/furiatona/azctl/discussions)
- 📧 **Email**: support@azctl.dev

### Troubleshooting

Common issues and solutions are documented in our [Troubleshooting Guide](TROUBLESHOOTING.md).

## 🙏 Acknowledgments

- Azure CLI team for the excellent Azure tooling
- Cobra team for the powerful CLI framework
- The Go community for the amazing ecosystem

---

**Made with ❤️ for the Azure community**

# Contributing to azctl

Thank you for your interest in contributing to azctl! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)
- [Contact](#contact)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- Azure CLI (for testing)
- Docker (for building images)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/azctl.git
   cd azctl
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/furiatona/azctl.git
   ```

## Development Setup

### Install Development Tools

```bash
make install-tools
```

This will install:
- `golangci-lint` for linting
- `goreleaser` for releases
- `goimports` for code formatting

### Build and Test

```bash
# Build the binary
make build

# Run tests
make test

# Run linting
make lint

# Run all checks
make ci
```

### Development Workflow

```bash
# Start development workflow
make dev
```

This will:
1. Download dependencies
2. Build the binary
3. Run tests
4. Show any issues

## Code Standards

### Go Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (run `make fmt`)
- Use `golangci-lint` for linting (run `make lint`)

### Naming Conventions

- Use `camelCase` for variables and functions
- Use `PascalCase` for exported types and functions
- Use `UPPER_SNAKE_CASE` for constants
- Use descriptive names that explain the purpose

### Error Handling

- Always check and handle errors explicitly
- Use `fmt.Errorf` with `%w` verb for wrapping errors
- Provide meaningful error messages
- Log errors appropriately

### Logging

- Use structured logging with the `logging` package
- Use appropriate log levels:
  - `Debug`: Detailed information for debugging
  - `Info`: General information about program execution
  - `Warn`: Warning messages for potentially harmful situations
  - `Error`: Error messages for error conditions

### Documentation

- Add comments for all exported functions and types
- Use [Go doc conventions](https://golang.org/doc/effective_go.html#commentary)
- Update README.md for user-facing changes
- Update inline documentation for code changes

### Configuration

- Use the `config` package for all configuration management
- Validate configuration using the `validation` package
- Provide sensible defaults where appropriate
- Support multiple configuration sources (env vars, files, etc.)

## Testing

### Test Structure

- Unit tests should be in `*_test.go` files
- Integration tests should use the `//go:build integration` tag
- Test files should be in the same package as the code they test

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run tests with coverage
make test
# Coverage report will be in coverage/coverage.html
```

### Writing Tests

- Aim for >90% test coverage
- Use table-driven tests for multiple test cases
- Mock external dependencies
- Test both success and failure scenarios
- Use descriptive test names

Example test structure:

```go
func TestFunctionName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid input",
			input:    "test",
			expected: "result",
			wantErr:  false,
		},
		{
			name:     "invalid input",
			input:    "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FunctionName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("FunctionName() = %v, want %v", result, tt.expected)
			}
		})
	}
}
```

## Pull Request Process

### Before Submitting

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write code following the standards above
   - Add tests for new functionality
   - Update documentation

3. **Run checks locally**:
   ```bash
   make ci
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

### Pull Request Guidelines

1. **Title**: Use conventional commit format (e.g., "feat: add new feature")
2. **Description**: Clearly describe what the PR does and why
3. **Tests**: Include tests for new functionality
4. **Documentation**: Update relevant documentation
5. **Breaking Changes**: Clearly mark and document any breaking changes

### Conventional Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

### Review Process

1. All PRs require at least one review
2. Maintainers will review for:
   - Code quality and standards
   - Test coverage
   - Documentation updates
   - Security considerations
3. Address review comments promptly
4. Keep PRs focused and reasonably sized

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/) (SemVer):

- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Creating a Release

1. **Update version**:
   ```bash
   # Update version in go.mod and other files
   make version
   ```

2. **Create release branch**:
   ```bash
   git checkout -b release/v1.2.3
   ```

3. **Run release checks**:
   ```bash
   make ci
   make release
   ```

4. **Create GitHub release**:
   - Tag the release
   - Upload binaries
   - Write release notes

### Release Notes

Include in release notes:
- New features
- Bug fixes
- Breaking changes
- Migration guide (if needed)
- Contributors

## Security

### Reporting Security Issues

- **DO NOT** create a public GitHub issue
- Email security issues to: security@azctl.dev
- Include detailed description and reproduction steps
- We will respond within 48 hours

### Security Guidelines

- Never commit secrets or sensitive data
- Use environment variables for configuration
- Validate all inputs
- Follow security best practices
- Keep dependencies updated

## Contact

- **Issues**: [GitHub Issues](https://github.com/furiatona/azctl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/furiatona/azctl/discussions)
- **Email**: support@azctl.dev
- **Security**: security@azctl.dev

## Acknowledgments

Thank you for contributing to azctl! Your contributions help make this tool better for everyone in the Azure community.

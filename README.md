# manifold-k8s

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidschrooten/manifold-k8s)](https://goreportcard.com/report/github.com/davidschrooten/manifold-k8s)
[![Test](https://github.com/davidschrooten/manifold-k8s/actions/workflows/test.yml/badge.svg)](https://github.com/davidschrooten/manifold-k8s/actions/workflows/test.yml)
[![Test Coverage](https://img.shields.io/badge/coverage-68.0%25-brightgreen)](https://github.com/davidschrooten/manifold-k8s)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A CLI tool that allows you to interactively or programmatically download Kubernetes manifests from one or multiple namespaces.

## Features

- 🔍 **Interactive Selection**: Choose clusters, namespaces, and resource types through intuitive prompts
- 📦 **Comprehensive Resource Support**: Downloads all resource types including Custom Resource Definitions (CRDs)
- ⎈ **Helm Values Export**: Export Helm release values from one or multiple namespaces
- 🚫 **Smart Filtering**: Automatically excludes PersistentVolumes and PersistentVolumeClaims
- 🧹 **Clean Manifests**: Removes runtime fields (status, managedFields, UIDs, etc.) for clean exports
- 📁 **Organized Output**: Manifests are organized by namespace/resource-type/name.yaml
- 🔄 **Multi-Cluster Support**: Export from multiple clusters in a single run
- 👁️ **Dry-Run Mode**: Preview what would be downloaded without writing files

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [GitHub Releases](https://github.com/davidschrooten/manifold-k8s/releases) page.

Available platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

After downloading, make the binary executable:
```bash
chmod +x manifold-k8s-*
sudo mv manifold-k8s-* /usr/local/bin/manifold-k8s
```

### From Source

```bash
git clone https://github.com/davidschrooten/manifold-k8s
cd manifold-k8s
make build
```

The binary will be available at `bin/manifold-k8s`.

### Using Go Install

```bash
go install github.com/davidschrooten/manifold-k8s@latest
```

## Usage

manifold-k8s provides several commands:
- `kubectl-manifests`: Interactive mode for Kubernetes manifests (best for exploration)
- `kubectl-manifests-export`: Non-interactive mode for Kubernetes manifests (best for scripting/CI-CD)
- `helm-values`: Interactive mode for Helm release values
- `helm-values-export`: Non-interactive mode for Helm release values

### Interactive Mode

```bash
manifold-k8s kubectl-manifests
```

This will start an interactive session that guides you through:
1. Selecting cluster context(s) from your kubeconfig
2. Selecting namespace(s)
3. Selecting resource type(s) to export
4. Choosing a target directory

### Export Mode (Non-Interactive)

```bash
manifold-k8s kubectl-manifests-export --context prod --namespaces default --resources pods,deployments -o ./output
```

This requires all parameters via flags (no prompts).

### Command Line Options

**Interactive Command:**
```bash
manifold-k8s kubectl-manifests [flags]

Flags:
      --dry-run         Preview what would be downloaded without writing files
  -o, --output string   Output directory (will be prompted if not provided)
```

**Export Command:**
```bash
manifold-k8s kubectl-manifests-export [flags]

Flags:
  -a, --all-resources        Export all resource types
  -c, --context string       Kubernetes context (required)
      --dry-run              Preview what would be exported without writing files
  -n, --namespaces strings   Namespaces to export (comma-separated, required)
  -o, --output string        Output directory (required)
  -r, --resources strings    Resource types to export (comma-separated, e.g. pods,deployments)
```

**Global Flags:**
```bash
  --kubeconfig string    Path to kubeconfig file (default is $HOME/.kube/config)
  --config string        Config file (default is ./config.toml)
```

### Examples

#### Interactive Command

**Basic interactive use:**
```bash
manifold-k8s kubectl-manifests
```

**Interactive with pre-specified output directory:**
```bash
manifold-k8s kubectl-manifests -o ./my-manifests
```

**Interactive dry-run:**
```bash
manifold-k8s kubectl-manifests --dry-run
```

**Use a specific kubeconfig:**
```bash
manifold-k8s kubectl-manifests --kubeconfig ~/.kube/config-prod
```

#### Export Command (Non-Interactive)

**Export specific resources from specific namespaces:**
```bash
manifold-k8s kubectl-manifests-export --context prod --namespaces default,kube-system --resources pods,deployments,services -o ./output
```

**Export all resources from a namespace:**
```bash
manifold-k8s kubectl-manifests-export --context staging --namespaces myapp --all-resources -o ./backup
```

**Dry-run in non-interactive mode:**
```bash
manifold-k8s kubectl-manifests-export --context prod --namespaces default --resources configmaps --dry-run -o ./test
```

**Export from multiple namespaces:**
```bash
manifold-k8s kubectl-manifests-export -c prod -n namespace1,namespace2,namespace3 -r deployments,statefulsets -o ./manifests
```

### Helm Values Export

Export Helm release values from your clusters. Requires `helm` CLI to be installed.

#### Interactive Helm Values Export

**Basic usage:**
```bash
manifold-k8s helm-values
```

This will interactively guide you through:
1. Selecting cluster context(s)
2. Selecting namespace(s)
3. Automatically discovers and exports all Helm releases

**With pre-specified output directory:**
```bash
manifold-k8s helm-values -o ./helm-backup
```

#### Non-Interactive Helm Values Export

**Export all Helm releases from specific namespaces:**
```bash
manifold-k8s helm-values-export --context prod --namespaces default,app1 --all -o ./helm-values
```

**Export specific Helm releases:**
```bash
manifold-k8s helm-values-export -c prod -n default -r myapp,nginx -o ./values
```

**Dry-run:**
```bash
manifold-k8s helm-values-export -c staging -n myapp --all --dry-run -o ./test
```

## Output Structure

Manifests are organized in the following directory structure:

```
output-directory/
├── namespace1/
│   ├── deployments/
│   │   ├── app1.yaml
│   │   └── app2.yaml
│   ├── services/
│   │   └── app1-svc.yaml
│   └── configmaps/
│       └── app-config.yaml
└── namespace2/
    └── pods/
        └── pod1.yaml
```

## Development

### Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster (for testing)

### Building

```bash
make build
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run linters
make lint
```

### Test Coverage

The project maintains high test coverage with extensive mocking:
- **Core Packages Average**: 91.7%
  - `pkg/exporter`: 100.0% (complete coverage with all error paths tested)
  - `pkg/k8s`: 92.5% (comprehensive resource discovery and filtering)
  - `pkg/selector`: 89.6% (extensive mocking of survey library)
  - `pkg/helm`: 84.8% (Helm CLI integration and parsing)
- **cmd Package**: 51.1%
  - Core functionality (non-interactive): ~85% coverage
  - Interactive functions excluded from coverage (require user input)
  - Includes comprehensive tests for:
    - Command structure and flags
    - Kubernetes manifest export logic
    - Helm values export logic
    - Error handling paths
- **Overall**: 68.0%

**Note:** Interactive command functions (`kubectl-manifests` and `helm-values`) are excluded from coverage metrics as they require user prompts and are not easily testable. All non-interactive command paths (`kubectl-manifests-export`, `helm-values-export`) have comprehensive test coverage.

### Development Workflow

The project follows Test-Driven Development (TDD) principles:
1. Write tests first
2. Implement functionality to make tests pass
3. Commit changes to feature branches
4. Merge to master after tests pass

### Releasing

Releases are automated via GitHub Actions. To create a new release:

1. Tag a commit with a version number:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. The release workflow will automatically:
   - Run tests
   - Build binaries for all supported platforms
   - Create checksums
   - Create a GitHub release with all artifacts
   - Generate release notes

## Architecture

The project is organized into three main packages:

- **pkg/k8s**: Kubernetes client management, resource discovery, and filtering
- **pkg/selector**: Interactive prompts using the survey library
- **pkg/exporter**: Manifest cleaning and file writing logic
- **cmd**: Cobra command structure and workflow orchestration

## Configuration

The tool uses Viper for configuration management. You can create a `config.toml` file:

```toml
kubeconfig = "/path/to/kubeconfig"
```

Or use environment variables with the `MANIFOLD_` prefix:

```bash
export MANIFOLD_KUBECONFIG=/path/to/kubeconfig
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please ensure:
- Tests are written for new functionality
- Test coverage remains high
- Code follows existing patterns and conventions
- Commit messages are descriptive

# manifold-k8s

A CLI tool that allows you to interactively select and download Kubernetes manifests from one or multiple namespaces.

## Features

- 🔍 **Interactive Selection**: Choose clusters, namespaces, and resource types through intuitive prompts
- 📦 **Comprehensive Resource Support**: Downloads all resource types including Custom Resource Definitions (CRDs)
- 🚫 **Smart Filtering**: Automatically excludes PersistentVolumes and PersistentVolumeClaims
- 🧹 **Clean Manifests**: Removes runtime fields (status, managedFields, UIDs, etc.) for clean exports
- 📁 **Organized Output**: Manifests are organized by namespace/resource-type/name.yaml
- 🔄 **Multi-Cluster Support**: Export from multiple clusters in a single run
- 👁️ **Dry-Run Mode**: Preview what would be downloaded without writing files

## Installation

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

### Basic Usage

```bash
manifold-k8s download
```

This will start an interactive session that guides you through:
1. Selecting cluster context(s) from your kubeconfig
2. Selecting namespace(s)
3. Selecting resource type(s) to export
4. Choosing a target directory

### Command Line Options

```bash
manifold-k8s download [flags]

Flags:
  -a, --all-resources        Export all resource types (non-interactive mode)
  -c, --context string       Kubernetes context (non-interactive mode)
      --dry-run              Preview what would be downloaded without writing files
  -n, --namespaces strings   Namespaces to export (comma-separated, non-interactive mode)
  -o, --output string        Output directory (will be prompted if not provided)
  -r, --resources strings    Resource types to export (comma-separated, e.g. pods,deployments)
      --kubeconfig string    Path to kubeconfig file (default is $HOME/.kube/config)
      --config string        Config file (default is ./config.toml)
```

### Examples

#### Interactive Mode

**Export from current cluster to a specific directory:**
```bash
manifold-k8s download -o ./my-manifests
```

**Preview what would be exported (dry-run):**
```bash
manifold-k8s download --dry-run
```

**Use a specific kubeconfig:**
```bash
manifold-k8s download --kubeconfig ~/.kube/config-prod
```

#### Non-Interactive Mode

**Export specific resources from specific namespaces:**
```bash
manifold-k8s download --context prod --namespaces default,kube-system --resources pods,deployments,services -o ./output
```

**Export all resources from a namespace:**
```bash
manifold-k8s download --context staging --namespaces myapp --all-resources -o ./backup
```

**Dry-run in non-interactive mode:**
```bash
manifold-k8s download --context prod --namespaces default --resources configmaps --dry-run -o ./test
```

**Export from multiple namespaces:**
```bash
manifold-k8s download -c prod -n namespace1,namespace2,namespace3 -r deployments,statefulsets -o ./manifests
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

The project maintains high test coverage:
- `pkg/k8s`: 87.5%
- `pkg/exporter`: 84.8%
- `pkg/selector`: Helper functions at 100% (interactive prompts require integration testing)

### Development Workflow

The project follows Test-Driven Development (TDD) principles:
1. Write tests first
2. Implement functionality to make tests pass
3. Commit changes to feature branches
4. Merge to master after tests pass

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

Co-Authored-By: Warp <agent@warp.dev>
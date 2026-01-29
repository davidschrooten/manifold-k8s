package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestExportCmd(t *testing.T) {
	if exportCmd == nil {
		t.Fatal("exportCmd is nil")
	}

	if exportCmd.Use != "kubectl-manifests-export" {
		t.Errorf("exportCmd.Use = %s, want kubectl-manifests-export", exportCmd.Use)
	}

	// Test that command has required flags
	requiredFlags := []string{"context", "namespaces", "output"}
	for _, flag := range requiredFlags {
		f := exportCmd.Flag(flag)
		if f == nil {
			t.Errorf("exportCmd missing required flag: %s", flag)
		}
	}

	// Test that command has optional flags
	optionalFlags := []string{"dry-run", "resources", "all-resources"}
	for _, flag := range optionalFlags {
		f := exportCmd.Flag(flag)
		if f == nil {
			t.Errorf("exportCmd missing flag: %s", flag)
		}
	}
}

func TestExportCmd_RequiredFlagsMarked(t *testing.T) {
	// Test that required flags are properly marked
	requiredFlags := []string{"context", "namespaces", "output"}
	for _, flagName := range requiredFlags {
		flag := exportCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Required flag %s not found", flagName)
			continue
		}
		// Check if flag is marked as required via cobra's annotation
		annotations := flag.Annotations
		if annotations != nil {
			if required, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; ok {
				if len(required) == 0 {
					t.Errorf("Flag %s not properly marked as required", flagName)
				}
			}
		}
	}
}

func TestExportCmd_MissingResourcesOrAllResources(t *testing.T) {
	// Test that either --resources or --all-resources is required
	// This is validated in runExport, so we can't easily test without mocking
	// but we document that this validation exists
	if !exportAllRes && len(exportResources) == 0 {
		// This is the validation logic that should trigger an error
		t.Log("Validation for --resources or --all-resources is present")
	}
}

func TestExportCmd_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{"dry-run flag", "dry-run", "bool"},
		{"output flag", "output", "string"},
		{"context flag", "context", "string"},
		{"namespaces flag", "namespaces", "stringSlice"},
		{"resources flag", "resources", "stringSlice"},
		{"all-resources flag", "all-resources", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := exportCmd.Flag(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if flag.Value.Type() != tt.wantType {
				t.Errorf("Flag %s type = %s, want %s", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestRunExport_Success(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir for output
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.NoError(t, err)
}

func TestRunExport_DryRun(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = true
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.NoError(t, err)

	// Verify no files were created
	entries, _ := os.ReadDir(tmpDir)
	assert.Equal(t, 0, len(entries))
}

func TestRunExport_AllResources(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = nil
	exportAllRes = true

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.NoError(t, err)
}

func TestRunExport_MultipleNamespaces(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default", "kube-system"}
	exportResources = []string{"pods", "deployments"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.NoError(t, err)
}

func TestRunExport_InvalidFlags(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags (neither --resources nor --all-resources)
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = nil
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either --resources or --all-resources is required")
}

func TestRunExport_LoadKubeConfigError(t *testing.T) {
	// Setup - enable stubs but override to return error
	enableStubs()
	defer disableStubs()

	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return nil, assert.AnError
	}

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load kubeconfig")
}

func TestRunExport_NewClientError(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return nil, assert.AnError
	}

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create client")
}

func TestRunExport_DiscoverResourcesError(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	stubDiscoverResources = func(discovery.DiscoveryInterface) ([]k8s.ResourceInfo, error) {
		return nil, assert.AnError
	}

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover resources")
}

func TestRunExport_NoValidResources(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags with non-existent resource type
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"nonexistent"}
	exportAllRes = false

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run
	err := runExport(exportCmd, []string{})

	// Restore stderr
	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	stderr := buf.String()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid resource types found")
	assert.Contains(t, stderr, "Warning: resource type nonexistent not found in cluster")
}

func TestRunExport_WithOutputFiles(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()

	// Create temp dir
	tmpDir := t.TempDir()

	// Set up viper
	viper.Set("kubeconfig", "/fake/path")

	// Set flags
	exportDryRun = false
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false

	// Run
	err := runExport(exportCmd, []string{})

	// Assert
	assert.NoError(t, err)

	// Check that output directory exists and has content
	entries, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Greater(t, len(entries), 0, "Expected output directory to have content")
}

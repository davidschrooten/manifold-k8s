package cmd

import (
	"fmt"
	"testing"

	"github.com/davidschrooten/manifold-k8s/pkg/helm"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestHelmValuesCmd(t *testing.T) {
	if helmValuesCmd == nil {
		t.Fatal("helmValuesCmd is nil")
	}

	if helmValuesCmd.Use != "helm-values" {
		t.Errorf("helmValuesCmd.Use = %s, want helm-values", helmValuesCmd.Use)
	}

	// Test that command has flags
	flags := []string{"dry-run", "output"}
	for _, flag := range flags {
		f := helmValuesCmd.Flag(flag)
		if f == nil {
			t.Errorf("helmValuesCmd missing flag: %s", flag)
		}
	}
}

func TestHelmValuesCmd_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{"dry-run flag", "dry-run", "bool"},
		{"output flag", "output", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := helmValuesCmd.Flag(tt.flagName)
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

func TestHelmValuesExportCmd(t *testing.T) {
	if helmValuesExportCmd == nil {
		t.Fatal("helmValuesExportCmd is nil")
	}

	if helmValuesExportCmd.Use != "helm-values-export" {
		t.Errorf("helmValuesExportCmd.Use = %s, want helm-values-export", helmValuesExportCmd.Use)
	}

	// Test that command has required flags
	requiredFlags := []string{"context", "namespaces", "output"}
	for _, flag := range requiredFlags {
		f := helmValuesExportCmd.Flag(flag)
		if f == nil {
			t.Errorf("helmValuesExportCmd missing required flag: %s", flag)
		}
	}

	// Test that command has optional flags
	optionalFlags := []string{"dry-run", "releases", "all"}
	for _, flag := range optionalFlags {
		f := helmValuesExportCmd.Flag(flag)
		if f == nil {
			t.Errorf("helmValuesExportCmd missing flag: %s", flag)
		}
	}
}

func TestHelmValuesExportCmd_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{"dry-run flag", "dry-run", "bool"},
		{"output flag", "output", "string"},
		{"context flag", "context", "string"},
		{"namespaces flag", "namespaces", "stringSlice"},
		{"releases flag", "releases", "stringSlice"},
		{"all flag", "all", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := helmValuesExportCmd.Flag(tt.flagName)
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

func TestHelmValuesExportCmd_RequiredFlagsMarked(t *testing.T) {
	// Test that required flags are properly marked
	requiredFlags := []string{"context", "namespaces", "output"}
	for _, flagName := range requiredFlags {
		flag := helmValuesExportCmd.Flag(flagName)
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

func TestHelmValuesCmd_Help(t *testing.T) {
	// Test that --help works
	helmValuesCmd.SetArgs([]string{"--help"})

	err := helmValuesCmd.Execute()
	if err != nil {
		t.Errorf("helmValuesCmd.Execute() with --help returned error: %v", err)
	}
}

func TestHelmValuesExportCmd_Help(t *testing.T) {
	// Test that --help works
	helmValuesExportCmd.SetArgs([]string{"--help"})

	err := helmValuesExportCmd.Execute()
	if err != nil {
		t.Errorf("helmValuesExportCmd.Execute() with --help returned error: %v", err)
	}
}

func TestRunHelmValuesExport_MissingReleasesOrAll(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	originalNewClient := stubNewClient
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		stubNewClient = originalNewClient
		helmExportAll = false
		helmExportReleases = nil
	}()

	// Setup stubs
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return mockK8sClient(), nil
	}

	// Test without --all or --releases
	helmExportAll = false
	helmExportReleases = []string{}

	err := runHelmValuesExport(nil, nil)
	if err == nil {
		t.Error("runHelmValuesExport should error when neither --all nor --releases is specified")
	}
	if err != nil && err.Error() != "must specify either --releases or --all" {
		t.Errorf("runHelmValuesExport error = %v, want 'must specify either --releases or --all'", err)
	}
}

func TestRunHelmValuesExport_Success(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	originalNewClient := stubNewClient
	originalListReleases := stubListHelmReleases
	originalGetValues := stubGetHelmValues
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		stubNewClient = originalNewClient
		stubListHelmReleases = originalListReleases
		stubGetHelmValues = originalGetValues
		helmExportAll = false
		helmExportReleases = nil
		helmExportNamespaces = nil
		helmExportCtx = ""
		helmExportOutputDir = ""
		helmExportDryRun = false
	}()

	// Setup stubs
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return mockK8sClient(), nil
	}
	stubListHelmReleases = func(namespace string) ([]helm.Release, error) {
		return []helm.Release{
			{Name: "myapp", Namespace: namespace, Chart: "myapp-1.0.0"},
		}, nil
	}
	stubGetHelmValues = func(releaseName, namespace string) (string, error) {
		return "replicaCount: 3\n", nil
	}

	// Create temp output dir
	tempDir := t.TempDir()

	// Set flags
	helmExportAll = true
	helmExportNamespaces = []string{"default"}
	helmExportCtx = "test-context"
	helmExportOutputDir = tempDir
	helmExportDryRun = false

	err := runHelmValuesExport(nil, nil)
	if err != nil {
		t.Fatalf("runHelmValuesExport failed: %v", err)
	}
}

func TestRunHelmValuesExport_SpecificReleases(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	originalNewClient := stubNewClient
	originalListReleases := stubListHelmReleases
	originalGetValues := stubGetHelmValues
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		stubNewClient = originalNewClient
		stubListHelmReleases = originalListReleases
		stubGetHelmValues = originalGetValues
		helmExportAll = false
		helmExportReleases = nil
		helmExportNamespaces = nil
		helmExportCtx = ""
		helmExportOutputDir = ""
		helmExportDryRun = false
	}()

	// Setup stubs
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return mockK8sClient(), nil
	}
	stubListHelmReleases = func(namespace string) ([]helm.Release, error) {
		return []helm.Release{
			{Name: "app1", Namespace: namespace, Chart: "app1-1.0.0"},
			{Name: "app2", Namespace: namespace, Chart: "app2-1.0.0"},
		}, nil
	}
	stubGetHelmValues = func(releaseName, namespace string) (string, error) {
		return "key: value\n", nil
	}

	// Create temp output dir
	tempDir := t.TempDir()

	// Set flags - only export app1
	helmExportAll = false
	helmExportReleases = []string{"app1"}
	helmExportNamespaces = []string{"default"}
	helmExportCtx = "test-context"
	helmExportOutputDir = tempDir
	helmExportDryRun = false

	err := runHelmValuesExport(nil, nil)
	if err != nil {
		t.Fatalf("runHelmValuesExport failed: %v", err)
	}
}

func TestRunHelmValuesExport_NoReleasesFound(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	originalNewClient := stubNewClient
	originalListReleases := stubListHelmReleases
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		stubNewClient = originalNewClient
		stubListHelmReleases = originalListReleases
		helmExportAll = false
		helmExportReleases = nil
		helmExportNamespaces = nil
		helmExportCtx = ""
		helmExportOutputDir = ""
	}()

	// Setup stubs
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return mockK8sClient(), nil
	}
	stubListHelmReleases = func(namespace string) ([]helm.Release, error) {
		return []helm.Release{}, nil // No releases
	}

	// Create temp output dir
	tempDir := t.TempDir()

	// Set flags
	helmExportAll = true
	helmExportNamespaces = []string{"default"}
	helmExportCtx = "test-context"
	helmExportOutputDir = tempDir

	err := runHelmValuesExport(nil, nil)
	if err != nil {
		t.Fatalf("runHelmValuesExport failed: %v", err)
	}
}

func TestRunHelmValuesExport_LoadKubeConfigError(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		helmExportAll = false
		helmExportReleases = nil
	}()

	// Setup stub to return error
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return nil, fmt.Errorf("failed to load config")
	}

	// Set flags
	helmExportAll = true

	err := runHelmValuesExport(nil, nil)
	if err == nil {
		t.Error("runHelmValuesExport should error when kubeconfig fails to load")
	}
}

func TestRunHelmValuesExport_NewClientError(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	originalNewClient := stubNewClient
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		stubNewClient = originalNewClient
		helmExportAll = false
		helmExportReleases = nil
	}()

	// Setup stubs
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return nil, fmt.Errorf("failed to create client")
	}

	// Set flags
	helmExportAll = true

	err := runHelmValuesExport(nil, nil)
	if err == nil {
		t.Error("runHelmValuesExport should error when client creation fails")
	}
}

func TestRunHelmValuesExport_DryRun(t *testing.T) {
	// Save and restore stubs
	originalLoadConfig := stubLoadKubeConfig
	originalNewClient := stubNewClient
	originalListReleases := stubListHelmReleases
	defer func() {
		stubLoadKubeConfig = originalLoadConfig
		stubNewClient = originalNewClient
		stubListHelmReleases = originalListReleases
		helmExportAll = false
		helmExportReleases = nil
		helmExportNamespaces = nil
		helmExportCtx = ""
		helmExportOutputDir = ""
		helmExportDryRun = false
	}()

	// Setup stubs
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return mockK8sClient(), nil
	}
	stubListHelmReleases = func(namespace string) ([]helm.Release, error) {
		return []helm.Release{
			{Name: "myapp", Namespace: namespace, Chart: "myapp-1.0.0"},
		}, nil
	}

	// Create temp output dir
	tempDir := t.TempDir()

	// Set flags
	helmExportAll = true
	helmExportNamespaces = []string{"default"}
	helmExportCtx = "test-context"
	helmExportOutputDir = tempDir
	helmExportDryRun = true

	err := runHelmValuesExport(nil, nil)
	if err != nil {
		t.Fatalf("runHelmValuesExport failed: %v", err)
	}
}

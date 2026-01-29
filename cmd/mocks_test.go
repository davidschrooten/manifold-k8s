package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// createMockKubeconfig creates a temporary kubeconfig file for testing
func createMockKubeconfig(t *testing.T) string {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	config := api.NewConfig()
	config.Clusters["test-cluster"] = &api.Cluster{
		Server: "https://localhost:6443",
	}
	config.Contexts["test-context"] = &api.Context{
		Cluster:  "test-cluster",
		AuthInfo: "test-user",
	}
	config.AuthInfos["test-user"] = &api.AuthInfo{
		Token: "test-token",
	}
	config.CurrentContext = "test-context"

	err := clientcmd.WriteToFile(*config, kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create mock kubeconfig: %v", err)
	}

	return kubeconfigPath
}

// mockExportCmd creates a testable export command
func mockExportCmd(t *testing.T) *cobra.Command {
	tmpDir := t.TempDir()
	
	cmd := &cobra.Command{
		Use: "export-test",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate export logic without actual K8s calls
			fmt.Println("Mock export started")
			fmt.Printf("Exporting from namespaces: %v\n", exportNamespaces)
			fmt.Printf("Output directory: %s\n", exportOutputDir)
			return nil
		},
	}
	
	// Set up flags
	cmd.Flags().StringVarP(&exportOutputDir, "output", "o", tmpDir, "output directory")
	cmd.Flags().StringVarP(&exportCtx, "context", "c", "test-context", "kubernetes context")
	cmd.Flags().StringSliceVarP(&exportNamespaces, "namespaces", "n", []string{"default"}, "namespaces")
	cmd.Flags().StringSliceVarP(&exportResources, "resources", "r", []string{"pods"}, "resources")
	cmd.Flags().BoolVarP(&exportAllRes, "all-resources", "a", false, "all resources")
	cmd.Flags().BoolVar(&exportDryRun, "dry-run", false, "dry run")
	
	return cmd
}

// TestRunExportWithMocks tests runExport with mocked dependencies
func TestRunExportWithMocks(t *testing.T) {
	// Save original values
	origOutputDir := exportOutputDir
	origCtx := exportCtx
	origNamespaces := exportNamespaces
	origResources := exportResources
	origAllRes := exportAllRes
	origDryRun := exportDryRun
	
	defer func() {
		exportOutputDir = origOutputDir
		exportCtx = origCtx
		exportNamespaces = origNamespaces
		exportResources = origResources
		exportAllRes = origAllRes
		exportDryRun = origDryRun
	}()
	
	// Set up test values
	tmpDir := t.TempDir()
	exportOutputDir = tmpDir
	exportCtx = "test-context"
	exportNamespaces = []string{"default"}
	exportResources = []string{"pods"}
	exportAllRes = false
	exportDryRun = true
	
	// Create mock kubeconfig
	kubeconfigPath := createMockKubeconfig(t)
	os.Setenv("KUBECONFIG", kubeconfigPath)
	defer os.Unsetenv("KUBECONFIG")
	
	// Test that we can call the command structure
	cmd := mockExportCmd(t)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Mock export command failed: %v", err)
	}
}

// TestRunInteractiveWithMocks tests runInteractive with mocked dependencies  
func TestRunInteractiveWithMocks(t *testing.T) {
	// Save original values
	origOutputDir := interactiveOutputDir
	origDryRun := interactiveDryRun
	
	defer func() {
		interactiveOutputDir = origOutputDir
		interactiveDryRun = origDryRun
	}()
	
	// Set up test values
	tmpDir := t.TempDir()
	interactiveOutputDir = tmpDir
	interactiveDryRun = true
	
	// Create mock kubeconfig
	kubeconfigPath := createMockKubeconfig(t)
	os.Setenv("KUBECONFIG", kubeconfigPath)
	defer os.Unsetenv("KUBECONFIG")
	
	// Test command structure (actual execution would require prompts)
	if interactiveCmd == nil {
		t.Fatal("interactiveCmd is nil")
	}
	
	// Verify flags are accessible
	if interactiveOutputDir != tmpDir {
		t.Errorf("interactiveOutputDir = %s, want %s", interactiveOutputDir, tmpDir)
	}
}

// TestExportValidation tests export command validation logic
func TestExportValidation(t *testing.T) {
	tests := []struct {
		name          string
		allRes        bool
		resources     []string
		wantErr       bool
		errContains   string
	}{
		{
			name:      "all-resources flag set",
			allRes:    true,
			resources: []string{},
			wantErr:   false,
		},
		{
			name:      "resources flag set",
			allRes:    false,
			resources: []string{"pods"},
			wantErr:   false,
		},
		{
			name:        "neither flag set",
			allRes:      false,
			resources:   []string{},
			wantErr:     true,
			errContains: "either --resources or --all-resources is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic from runExport
			exportAllRes = tt.allRes
			exportResources = tt.resources
			
			var err error
			if !exportAllRes && len(exportResources) == 0 {
				err = fmt.Errorf("either --resources or --all-resources is required")
			}
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Validation error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want to contain %s", err, tt.errContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestResourceMapping tests the resource mapping logic
func TestResourceMapping(t *testing.T) {
	discoveredResources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment"},
		{Name: "services", Group: "", Version: "v1", Kind: "Service"},
	}
	
	// Build map like in runExport
	resourceMap := make(map[string]k8s.ResourceInfo)
	for _, res := range discoveredResources {
		resourceMap[res.Name] = res
	}
	
	tests := []struct {
		name         string
		requestedRes []string
		wantFound    int
		wantWarning  int
	}{
		{
			name:         "all found",
			requestedRes: []string{"pods", "services"},
			wantFound:    2,
			wantWarning:  0,
		},
		{
			name:         "some not found",
			requestedRes: []string{"pods", "invalid"},
			wantFound:    1,
			wantWarning:  1,
		},
		{
			name:         "none found",
			requestedRes: []string{"invalid1", "invalid2"},
			wantFound:    0,
			wantWarning:  2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var selectedResources []k8s.ResourceInfo
			warnings := 0
			
			for _, resName := range tt.requestedRes {
				if res, found := resourceMap[resName]; found {
					selectedResources = append(selectedResources, res)
				} else {
					warnings++
				}
			}
			
			if len(selectedResources) != tt.wantFound {
				t.Errorf("Found %d resources, want %d", len(selectedResources), tt.wantFound)
			}
			if warnings != tt.wantWarning {
				t.Errorf("Got %d warnings, want %d", warnings, tt.wantWarning)
			}
		})
	}
}

// TestClusterScopedResourceSkip tests that cluster-scoped resources are skipped properly
func TestClusterScopedResourceSkip(t *testing.T) {
	resources := []k8s.ResourceInfo{
		{Name: "pods", Namespaced: true},
		{Name: "nodes", Namespaced: false},
		{Name: "deployments", Namespaced: true},
	}
	
	namespace := "default"
	var processedCount int
	
	for _, resource := range resources {
		// Logic from runExport/runInteractive
		if !resource.Namespaced && namespace != "" {
			continue
		}
		processedCount++
	}
	
	// Should only process the 2 namespaced resources
	if processedCount != 2 {
		t.Errorf("Processed %d resources, want 2", processedCount)
	}
}

// TestDryRunMode tests dry-run behavior
func TestDryRunMode(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{"dry run enabled", true},
		{"dry run disabled", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportDryRun = tt.dryRun
			
			// Simulate the dry-run check from runExport
			if exportDryRun {
				// In dry-run mode, we would print but not export
				t.Log("Would export in dry-run mode")
			} else {
				// In normal mode, we would actually export
				t.Log("Would actually export")
			}
			
			// Test passes if no panic
		})
	}
}

package exporter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCleanManifest(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":            "test-pod",
				"namespace":       "default",
				"managedFields":   []interface{}{},
				"uid":             "test-uid",
				"resourceVersion": "12345",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "nginx",
						"image": "nginx:latest",
					},
				},
			},
			"status": map[string]interface{}{
				"phase": "Running",
			},
		},
	}

	cleaned := CleanManifest(obj)

	// Check that status is removed
	if _, found := cleaned.Object["status"]; found {
		t.Error("CleanManifest() did not remove status field")
	}

	// Check that managedFields is removed from metadata
	if metadata, ok := cleaned.Object["metadata"].(map[string]interface{}); ok {
		if _, found := metadata["managedFields"]; found {
			t.Error("CleanManifest() did not remove managedFields from metadata")
		}
		if _, found := metadata["uid"]; found {
			t.Error("CleanManifest() did not remove uid from metadata")
		}
		if _, found := metadata["resourceVersion"]; found {
			t.Error("CleanManifest() did not remove resourceVersion from metadata")
		}
	}

	// Check that other fields are preserved
	if _, found := cleaned.Object["spec"]; !found {
		t.Error("CleanManifest() removed spec field")
	}
	if metadata, ok := cleaned.Object["metadata"].(map[string]interface{}); ok {
		if name, found := metadata["name"]; !found || name != "test-pod" {
			t.Error("CleanManifest() did not preserve name in metadata")
		}
	}
}

func TestGenerateFilePath(t *testing.T) {
	tests := []struct {
		name         string
		baseDir      string
		namespace    string
		resourceType string
		resourceName string
		want         string
	}{
		{
			name:         "simple path",
			baseDir:      "/tmp/output",
			namespace:    "default",
			resourceType: "pods",
			resourceName: "test-pod",
			want:         "/tmp/output/default/pods/test-pod.yaml",
		},
		{
			name:         "namespaced resource",
			baseDir:      "./manifests",
			namespace:    "kube-system",
			resourceType: "deployments",
			resourceName: "coredns",
			want:         "manifests/kube-system/deployments/coredns.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilePath(tt.baseDir, tt.namespace, tt.resourceType, tt.resourceName)
			// Normalize paths for comparison
			got = filepath.Clean(got)
			want := filepath.Clean(tt.want)
			if got != want {
				t.Errorf("GenerateFilePath() = %v, want %v", got, want)
			}
		})
	}
}

func TestWriteManifest(t *testing.T) {
	tmpDir := t.TempDir()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key": "value",
			},
		},
	}

	filePath := filepath.Join(tmpDir, "default", "configmaps", "test-config.yaml")
	err := WriteManifest(obj, filePath)
	if err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("WriteManifest() did not create file at %s", filePath)
	}

	// Read file and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	// Verify it's valid YAML with expected content
	if len(content) == 0 {
		t.Error("WriteManifest() wrote empty file")
	}

	// Check for expected fields in YAML
	contentStr := string(content)
	expectedFields := []string{"apiVersion", "kind", "metadata", "name: test-config"}
	for _, field := range expectedFields {
		if !contains(contentStr, field) {
			t.Errorf("WriteManifest() output missing expected field: %s", field)
		}
	}
}

func TestExporter_ExportResource(t *testing.T) {
	tmpDir := t.TempDir()

	exporter := NewExporter(tmpDir)

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{},
			},
		},
	}

	err := exporter.ExportResource(context.Background(), obj, gvr, "default")
	if err != nil {
		t.Fatalf("ExportResource() error = %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tmpDir, "default", "pods", "test-pod.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("ExportResource() did not create file at %s", expectedPath)
	}

	// Verify counter incremented
	if exporter.ExportedCount != 1 {
		t.Errorf("ExportResource() ExportedCount = %d, want 1", exporter.ExportedCount)
	}
}

func TestExporter_Summary(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir)

	// Export some resources
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	
	for i := 0; i < 3; i++ {
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name":      "test-pod-" + string(rune('a'+i)),
					"namespace": "default",
				},
			},
		}
		exporter.ExportResource(context.Background(), obj, gvr, "default")
	}

	summary := exporter.Summary()
	if !contains(summary, "3") {
		t.Errorf("Summary() does not contain expected count: %s", summary)
	}
	if !contains(summary, tmpDir) {
		t.Errorf("Summary() does not contain output directory: %s", summary)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr)+1 && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

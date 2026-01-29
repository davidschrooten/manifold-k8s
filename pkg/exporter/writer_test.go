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
		_ = exporter.ExportResource(context.Background(), obj, gvr, "default")
	}

	summary := exporter.Summary()
	if !contains(summary, "3") {
		t.Errorf("Summary() does not contain expected count: %s", summary)
	}
	if !contains(summary, tmpDir) {
		t.Errorf("Summary() does not contain output directory: %s", summary)
	}
}

func TestWriteManifest_DirectoryError(t *testing.T) {
	// Use an invalid directory path that cannot be created
	invalidPath := "/root/nonexistent/test.yaml"

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
		},
	}

	err := WriteManifest(obj, invalidPath)
	if err == nil {
		t.Error("WriteManifest() expected error for invalid path, got nil")
	}
}

func TestExportResource_WriteError(t *testing.T) {
	// Use an invalid output directory
	exporter := NewExporter("/root/nonexistent")

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
		},
	}

	err := exporter.ExportResource(context.Background(), obj, gvr, "default")
	if err == nil {
		t.Error("ExportResource() expected error for invalid path, got nil")
	}

	// Verify counter did not increment on error
	if exporter.ExportedCount != 0 {
		t.Errorf("ExportResource() ExportedCount = %d, want 0 after error", exporter.ExportedCount)
	}
}

func TestExportResource_EmptyName(t *testing.T) {
	exporter := NewExporter(t.TempDir())

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// Object with no name
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"namespace": "default",
			},
		},
	}

	err := exporter.ExportResource(context.Background(), obj, gvr, "default")
	if err == nil {
		t.Error("ExportResource() expected error for empty name, got nil")
	}
	if !contains(err.Error(), "name") {
		t.Errorf("ExportResource() error should mention name, got: %v", err)
	}
}

func TestWriteManifest_WriteFileError(t *testing.T) {
	// Create a file where we want to create a directory
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "blockingfile")
	if err := os.WriteFile(blockingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to write to a path that requires creating a directory where a file exists
	invalidPath := filepath.Join(blockingFile, "subdir", "test.yaml")

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
		},
	}

	err := WriteManifest(obj, invalidPath)
	if err == nil {
		t.Error("WriteManifest() expected error when directory creation fails, got nil")
	}
}

func TestWriteManifest_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.yaml")

	// Create an object with an invalid value that can't be marshaled to YAML
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"invalid":    make(chan int), // channels can't be marshaled to YAML
		},
	}

	err := WriteManifest(obj, filePath)
	if err == nil {
		t.Error("WriteManifest() expected error for invalid YAML, got nil")
	}
	if !contains(err.Error(), "marshal") {
		t.Errorf("WriteManifest() error should mention marshal, got: %v", err)
	}
}

func TestWriteManifest_FileWriteError(t *testing.T) {
	// Create a read-only directory
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(readOnlyDir, 0555); err != nil { // read + execute only
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }() // cleanup

	filePath := filepath.Join(readOnlyDir, "test.yaml")

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
		},
	}

	err := WriteManifest(obj, filePath)
	if err == nil {
		t.Error("WriteManifest() expected error for read-only directory, got nil")
	}
	if !contains(err.Error(), "write file") && !contains(err.Error(), "failed to write") {
		t.Errorf("WriteManifest() error should mention file write, got: %v", err)
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

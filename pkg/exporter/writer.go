package exporter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// Exporter handles exporting Kubernetes manifests to disk
type Exporter struct {
	BaseDir       string
	ExportedCount int
	mu            sync.Mutex
}

// NewExporter creates a new Exporter
func NewExporter(baseDir string) *Exporter {
	return &Exporter{
		BaseDir:       baseDir,
		ExportedCount: 0,
	}
}

// CleanManifest removes runtime fields from a manifest
func CleanManifest(obj *unstructured.Unstructured) *unstructured.Unstructured {
	cleaned := obj.DeepCopy()

	// Remove status
	delete(cleaned.Object, "status")

	// Clean metadata
	if metadata, ok := cleaned.Object["metadata"].(map[string]interface{}); ok {
		// Remove managed fields and other runtime metadata
		delete(metadata, "managedFields")
		delete(metadata, "uid")
		delete(metadata, "resourceVersion")
		delete(metadata, "generation")
		delete(metadata, "creationTimestamp")
		delete(metadata, "selfLink")
	}

	return cleaned
}

// GenerateFilePath generates the file path for a resource
func GenerateFilePath(baseDir, namespace, resourceType, resourceName string) string {
	return filepath.Join(baseDir, namespace, resourceType, resourceName+".yaml")
}

// WriteManifest writes a manifest to a file
func WriteManifest(obj *unstructured.Unstructured, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ExportResource exports a single resource to disk
func (e *Exporter) ExportResource(ctx context.Context, obj *unstructured.Unstructured, gvr schema.GroupVersionResource, namespace string) error {
	// Clean the manifest
	cleaned := CleanManifest(obj)

	// Get resource name
	name := cleaned.GetName()
	if name == "" {
		return fmt.Errorf("resource has no name")
	}

	// Generate file path
	filePath := GenerateFilePath(e.BaseDir, namespace, gvr.Resource, name)

	// Write manifest
	if err := WriteManifest(cleaned, filePath); err != nil {
		return err
	}

	// Increment counter
	e.mu.Lock()
	e.ExportedCount++
	e.mu.Unlock()

	return nil
}

// Summary returns a summary of the export
func (e *Exporter) Summary() string {
	return fmt.Sprintf("Exported %d manifests to %s", e.ExportedCount, e.BaseDir)
}

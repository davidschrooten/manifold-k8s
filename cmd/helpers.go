package cmd

import (
	"fmt"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
)

// validateExportFlags validates export command flags
func validateExportFlags(allRes bool, resources []string) error {
	if !allRes && len(resources) == 0 {
		return fmt.Errorf("either --resources or --all-resources is required")
	}
	return nil
}

// buildResourceMap creates a map of resource names to ResourceInfo
func buildResourceMap(resources []k8s.ResourceInfo) map[string]k8s.ResourceInfo {
	resourceMap := make(map[string]k8s.ResourceInfo)
	for _, res := range resources {
		resourceMap[res.Name] = res
	}
	return resourceMap
}

// selectRequestedResources filters resources based on requested names
func selectRequestedResources(resourceMap map[string]k8s.ResourceInfo, requested []string) ([]k8s.ResourceInfo, []string) {
	var selected []k8s.ResourceInfo
	var notFound []string

	for _, resName := range requested {
		if res, found := resourceMap[resName]; found {
			selected = append(selected, res)
		} else {
			notFound = append(notFound, resName)
		}
	}

	return selected, notFound
}

// shouldProcessResource determines if a resource should be processed for a namespace
func shouldProcessResource(resource k8s.ResourceInfo, namespace string) bool {
	// Skip cluster-scoped resources when processing namespaces
	if !resource.Namespaced && namespace != "" {
		return false
	}
	return true
}

// formatOutputMessage returns a formatted output message
func formatOutputMessage(dryRun bool, namespace, resourceType, resourceName string) string {
	if dryRun {
		return fmt.Sprintf("[DRY-RUN] Would export: %s/%s/%s", namespace, resourceType, resourceName)
	}
	return fmt.Sprintf("Exported: %s/%s/%s", namespace, resourceType, resourceName)
}

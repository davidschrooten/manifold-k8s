package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

// ResourceInfo contains information about a Kubernetes resource type
type ResourceInfo struct {
	Name       string
	Group      string
	Version    string
	Kind       string
	Namespaced bool
}

// String returns a formatted string representation of the resource
func (r ResourceInfo) String() string {
	if r.Group == "" {
		return fmt.Sprintf("%s (%s/%s)", r.Name, r.Version, r.Kind)
	}
	return fmt.Sprintf("%s (%s/%s/%s)", r.Name, r.Group, r.Version, r.Kind)
}

// GroupVersionResource returns the GVR for this resource
func (r ResourceInfo) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    r.Group,
		Version:  r.Version,
		Resource: r.Name,
	}
}

// excludedResources is a list of resource types to exclude
var excludedResources = map[string]bool{
	"persistentvolumes":      true,
	"persistentvolumeclaims": true,
}

// priorityResources defines the display order for common resources
var priorityResources = []string{
	"deployments",
	"statefulsets",
	"daemonsets",
	"replicasets",
	"ingresses",
	"services",
	"configmaps",
	"secrets",
	"serviceaccounts",
	"pods",
	"jobs",
	"cronjobs",
	"persistentvolumeclaims",
}

// shouldExcludeResource checks if a resource should be excluded
func shouldExcludeResource(resourceName string) bool {
	return excludedResources[resourceName]
}

// DiscoverResources discovers all available Kubernetes resources
// It excludes PersistentVolumes and PersistentVolumeClaims
func DiscoverResources(discoveryClient discovery.DiscoveryInterface) ([]ResourceInfo, error) {
	// Get all API resource lists
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		// Ignore partial errors, as some API groups may not be available
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("failed to discover resources: %w", err)
		}
	}

	var resources []ResourceInfo
	for _, apiResourceList := range apiResourceLists {
		// Parse group version
		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			// Skip subresources (e.g., pods/status, pods/log)
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			// Skip excluded resources
			if shouldExcludeResource(apiResource.Name) {
				continue
			}

			// Only include resources that support list and get
			hasList := false
			hasGet := false
			for _, verb := range apiResource.Verbs {
				if verb == "list" {
					hasList = true
				}
				if verb == "get" {
					hasGet = true
				}
			}
			if !hasList || !hasGet {
				continue
			}

			resources = append(resources, ResourceInfo{
				Name:       apiResource.Name,
				Group:      gv.Group,
				Version:    gv.Version,
				Kind:       apiResource.Kind,
				Namespaced: apiResource.Namespaced,
			})
		}
	}

	// Sort resources by priority
	sortResourcesByPriority(resources)

	return resources, nil
}

// sortResourcesByPriority sorts resources in the following order:
// 1. Priority resources (in defined order)
// 2. Core/standard resources (alphabetically)
// 3. Custom resources (CRDs) (alphabetically)
func sortResourcesByPriority(resources []ResourceInfo) {
	// Create priority map for O(1) lookup
	priorityMap := make(map[string]int)
	for i, name := range priorityResources {
		priorityMap[name] = i
	}

	sort.SliceStable(resources, func(i, j int) bool {
		resI := resources[i]
		resJ := resources[j]

		// Check if either resource is in priority list
		prioI, hasPrioI := priorityMap[resI.Name]
		prioJ, hasPrioJ := priorityMap[resJ.Name]

		if hasPrioI && hasPrioJ {
			// Both are priority resources, sort by priority order
			return prioI < prioJ
		}
		if hasPrioI {
			// Only i is priority, it comes first
			return true
		}
		if hasPrioJ {
			// Only j is priority, it comes first
			return false
		}

		// Neither is priority resource
		// Check if they are core/standard vs custom resources
		isCustomI := isCustomResource(resI)
		isCustomJ := isCustomResource(resJ)

		if isCustomI && !isCustomJ {
			// i is custom, j is standard - j comes first
			return false
		}
		if !isCustomI && isCustomJ {
			// i is standard, j is custom - i comes first
			return true
		}

		// Both are same category (both standard or both custom), sort alphabetically
		return resI.Name < resJ.Name
	})
}

// isCustomResource determines if a resource is a custom resource (CRD)
func isCustomResource(res ResourceInfo) bool {
	// Core API resources have empty group or well-known groups
	if res.Group == "" {
		return false
	}

	// Well-known Kubernetes API groups (not custom)
	standardGroups := []string{
		"apps",
		"batch",
		"autoscaling",
		"policy",
		"rbac.authorization.k8s.io",
		"networking.k8s.io",
		"storage.k8s.io",
		"apiextensions.k8s.io",
		"admissionregistration.k8s.io",
		"scheduling.k8s.io",
		"coordination.k8s.io",
		"node.k8s.io",
		"discovery.k8s.io",
		"flowcontrol.apiserver.k8s.io",
		"certificates.k8s.io",
	}

	for _, sg := range standardGroups {
		if res.Group == sg {
			return false
		}
	}

	// If not in standard groups, it's a custom resource
	return true
}

// GetNamespaces retrieves all namespaces from the cluster
func GetNamespaces(ctx context.Context, client *Client) ([]string, error) {
	namespaceList, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

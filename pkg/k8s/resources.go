package k8s

import (
	"context"
	"fmt"
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

	return resources, nil
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

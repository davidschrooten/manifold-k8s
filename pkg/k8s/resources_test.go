package k8s

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestShouldExcludeResource(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		want     bool
	}{
		{
			name:     "exclude persistentvolumes",
			resource: "persistentvolumes",
			want:     true,
		},
		{
			name:     "exclude persistentvolumeclaims",
			resource: "persistentvolumeclaims",
			want:     true,
		},
		{
			name:     "include deployments",
			resource: "deployments",
			want:     false,
		},
		{
			name:     "include services",
			resource: "services",
			want:     false,
		},
		{
			name:     "include custom resources",
			resource: "myresources",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldExcludeResource(tt.resource)
			if got != tt.want {
				t.Errorf("shouldExcludeResource(%s) = %v, want %v", tt.resource, got, tt.want)
			}
		})
	}
}

func TestDiscoverResources(t *testing.T) {
	// Create a fake clientset with API resources
	fakeClient := fake.NewSimpleClientset() //nolint:staticcheck // Using deprecated API for testing purposes
	fakeDiscovery := fakeClient.Discovery().(*fakediscovery.FakeDiscovery)

	// Set up fake API resources
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"list", "get"}},
				{Name: "services", Namespaced: true, Kind: "Service", Verbs: []string{"list", "get"}},
				{Name: "persistentvolumes", Namespaced: false, Kind: "PersistentVolume", Verbs: []string{"list", "get"}},
				{Name: "persistentvolumeclaims", Namespaced: true, Kind: "PersistentVolumeClaim", Verbs: []string{"list", "get"}},
			},
		},
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: []string{"list", "get"}},
				{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet", Verbs: []string{"list", "get"}},
			},
		},
	}

	resources, err := DiscoverResources(fakeDiscovery)
	if err != nil {
		t.Fatalf("DiscoverResources() error = %v", err)
	}

	// Should exclude PV and PVC
	if len(resources) != 4 {
		t.Errorf("DiscoverResources() returned %d resources, want 4", len(resources))
	}

	// Verify resources are correct
	expectedResources := map[string]bool{
		"pods":         true,
		"services":     true,
		"deployments":  true,
		"statefulsets": true,
	}

	for _, res := range resources {
		if !expectedResources[res.Name] {
			t.Errorf("DiscoverResources() returned unexpected resource: %s", res.Name)
		}
	}

	// Verify PV and PVC are excluded
	for _, res := range resources {
		if res.Name == "persistentvolumes" || res.Name == "persistentvolumeclaims" {
			t.Errorf("DiscoverResources() should exclude %s", res.Name)
		}
	}
}

func TestGetNamespaces(t *testing.T) {
	// Create a fake clientset
	fakeClient := fake.NewSimpleClientset() //nolint:staticcheck // Using deprecated API for testing purposes

	// Create some namespaces
	ctx := context.Background()
	namespaces := []string{"default", "kube-system", "test-ns"}
	for _, ns := range namespaces {
		_, _ = fakeClient.CoreV1().Namespaces().Create(ctx, v1Namespace(ns), metav1.CreateOptions{})
	}

	client := &Client{
		Clientset: fakeClient,
	}

	result, err := GetNamespaces(ctx, client)
	if err != nil {
		t.Fatalf("GetNamespaces() error = %v", err)
	}

	if len(result) != len(namespaces) {
		t.Errorf("GetNamespaces() returned %d namespaces, want %d", len(result), len(namespaces))
	}

	expectedNs := map[string]bool{"default": true, "kube-system": true, "test-ns": true}
	for _, ns := range result {
		if !expectedNs[ns] {
			t.Errorf("GetNamespaces() returned unexpected namespace: %s", ns)
		}
	}
}

func TestResourceInfo_String(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceInfo
		want     string
	}{
		{
			name: "core resource",
			resource: ResourceInfo{
				Name:       "pods",
				Group:      "",
				Version:    "v1",
				Kind:       "Pod",
				Namespaced: true,
			},
			want: "pods (v1/Pod)",
		},
		{
			name: "apps resource",
			resource: ResourceInfo{
				Name:       "deployments",
				Group:      "apps",
				Version:    "v1",
				Kind:       "Deployment",
				Namespaced: true,
			},
			want: "deployments (apps/v1/Deployment)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.String()
			if got != tt.want {
				t.Errorf("ResourceInfo.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceInfo_GroupVersionResource(t *testing.T) {
	resource := ResourceInfo{
		Name:    "deployments",
		Group:   "apps",
		Version: "v1",
	}

	gvr := resource.GroupVersionResource()

	expected := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	if gvr != expected {
		t.Errorf("GroupVersionResource() = %v, want %v", gvr, expected)
	}
}

func TestIsCustomResource(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceInfo
		want     bool
	}{
		{
			name:     "core resource (empty group)",
			resource: ResourceInfo{Name: "pods", Group: "", Version: "v1"},
			want:     false,
		},
		{
			name:     "apps group resource",
			resource: ResourceInfo{Name: "deployments", Group: "apps", Version: "v1"},
			want:     false,
		},
		{
			name:     "batch group resource",
			resource: ResourceInfo{Name: "jobs", Group: "batch", Version: "v1"},
			want:     false,
		},
		{
			name:     "networking.k8s.io resource",
			resource: ResourceInfo{Name: "ingresses", Group: "networking.k8s.io", Version: "v1"},
			want:     false,
		},
		{
			name:     "custom resource",
			resource: ResourceInfo{Name: "mycustomresources", Group: "example.com", Version: "v1alpha1"},
			want:     true,
		},
		{
			name:     "another custom resource",
			resource: ResourceInfo{Name: "foos", Group: "mycompany.io", Version: "v1"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCustomResource(tt.resource)
			if got != tt.want {
				t.Errorf("isCustomResource(%v) = %v, want %v", tt.resource.Name, got, tt.want)
			}
		})
	}
}

func TestSortResourcesByPriority(t *testing.T) {
	resources := []ResourceInfo{
		{Name: "customresources", Group: "example.com", Version: "v1"},
		{Name: "services", Group: "", Version: "v1"},
		{Name: "configmaps", Group: "", Version: "v1"},
		{Name: "deployments", Group: "apps", Version: "v1"},
		{Name: "anothercrd", Group: "mycompany.io", Version: "v1"},
		{Name: "pods", Group: "", Version: "v1"},
		{Name: "statefulsets", Group: "apps", Version: "v1"},
		{Name: "namespaces", Group: "", Version: "v1"},
	}

	sortResourcesByPriority(resources)

	// Check priority resources come first in order
	if resources[0].Name != "deployments" {
		t.Errorf("First resource should be deployments, got %s", resources[0].Name)
	}
	if resources[1].Name != "statefulsets" {
		t.Errorf("Second resource should be statefulsets, got %s", resources[1].Name)
	}
	if resources[2].Name != "services" {
		t.Errorf("Third resource should be services, got %s", resources[2].Name)
	}
	if resources[3].Name != "configmaps" {
		t.Errorf("Fourth resource should be configmaps, got %s", resources[3].Name)
	}
	if resources[4].Name != "pods" {
		t.Errorf("Fifth resource should be pods, got %s", resources[4].Name)
	}

	// Check standard non-priority resources come before CRDs
	if resources[5].Name != "namespaces" {
		t.Errorf("Sixth resource should be namespaces (standard), got %s", resources[5].Name)
	}

	// Check CRDs come last (sorted alphabetically)
	if resources[6].Name != "anothercrd" {
		t.Errorf("Seventh resource should be anothercrd (CRD), got %s", resources[6].Name)
	}
	if resources[7].Name != "customresources" {
		t.Errorf("Eighth resource should be customresources (CRD), got %s", resources[7].Name)
	}
}

func TestDiscoverResources_SubresourcesFiltered(t *testing.T) {
	fakeClient := fake.NewSimpleClientset() //nolint:staticcheck // Using deprecated API for testing purposes
	fakeDiscovery := fakeClient.Discovery().(*fakediscovery.FakeDiscovery)

	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"list", "get"}},
				{Name: "pods/status", Namespaced: true, Kind: "Pod", Verbs: []string{"get", "patch"}},
				{Name: "pods/log", Namespaced: true, Kind: "Pod", Verbs: []string{"get"}},
			},
		},
	}

	resources, err := DiscoverResources(fakeDiscovery)
	if err != nil {
		t.Fatalf("DiscoverResources() error = %v", err)
	}

	// Should only return "pods", not subresources
	if len(resources) != 1 {
		t.Errorf("DiscoverResources() returned %d resources, want 1", len(resources))
	}
	if resources[0].Name != "pods" {
		t.Errorf("DiscoverResources() returned %s, want pods", resources[0].Name)
	}
}

func TestDiscoverResources_MissingVerbs(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	fakeDiscovery := fakeClient.Discovery().(*fakediscovery.FakeDiscovery)

	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"list", "get"}},
				{Name: "services", Namespaced: true, Kind: "Service", Verbs: []string{"list"}},    // missing 'get'
				{Name: "configmaps", Namespaced: true, Kind: "ConfigMap", Verbs: []string{"get"}}, // missing 'list'
			},
		},
	}

	resources, err := DiscoverResources(fakeDiscovery)
	if err != nil {
		t.Fatalf("DiscoverResources() error = %v", err)
	}

	// Should only return "pods" since it has both list and get
	if len(resources) != 1 {
		t.Errorf("DiscoverResources() returned %d resources, want 1", len(resources))
	}
	if resources[0].Name != "pods" {
		t.Errorf("DiscoverResources() returned %s, want pods", resources[0].Name)
	}
}

func TestSortResourcesByPriority_AllCategories(t *testing.T) {
	resources := []ResourceInfo{
		{Name: "zzz-custom", Group: "example.com", Version: "v1"},
		{Name: "aaa-custom", Group: "mycompany.io", Version: "v1"},
		{Name: "services", Group: "", Version: "v1"},
		{Name: "deployments", Group: "apps", Version: "v1"},
		{Name: "roles", Group: "rbac.authorization.k8s.io", Version: "v1"},
		{Name: "zzz-standard", Group: "apps", Version: "v1"},
		{Name: "aaa-standard", Group: "batch", Version: "v1"},
	}

	sortResourcesByPriority(resources)

	// Priority resource should be first
	if resources[0].Name != "deployments" {
		t.Errorf("First should be deployments (priority), got %s", resources[0].Name)
	}
	if resources[1].Name != "services" {
		t.Errorf("Second should be services (priority), got %s", resources[1].Name)
	}

	// Standard resources (not in priority, not custom) should come next, alphabetically
	foundStandardSection := false
	for i := 2; i < len(resources)-2; i++ {
		if resources[i].Name == "aaa-standard" || resources[i].Name == "roles" || resources[i].Name == "zzz-standard" {
			foundStandardSection = true
		}
	}
	if !foundStandardSection {
		t.Error("Standard resources should be in the middle section")
	}

	// CRDs should be last, alphabetically
	lastTwo := resources[len(resources)-2:]
	if lastTwo[0].Name != "aaa-custom" || lastTwo[1].Name != "zzz-custom" {
		t.Errorf("Last two should be custom resources alphabetically (aaa-custom, zzz-custom), got %s, %s",
			lastTwo[0].Name, lastTwo[1].Name)
	}
}

// Helper function to create namespace object
func v1Namespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

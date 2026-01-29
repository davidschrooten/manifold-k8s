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
	fakeClient := fake.NewSimpleClientset()
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
	fakeClient := fake.NewSimpleClientset()

	// Create some namespaces
	ctx := context.Background()
	namespaces := []string{"default", "kube-system", "test-ns"}
	for _, ns := range namespaces {
		fakeClient.CoreV1().Namespaces().Create(ctx, v1Namespace(ns), metav1.CreateOptions{})
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

// Helper function to create namespace object
func v1Namespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

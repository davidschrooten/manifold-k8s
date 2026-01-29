package cmd

import (
	"testing"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
)

func TestValidateExportFlags(t *testing.T) {
	tests := []struct {
		name      string
		allRes    bool
		resources []string
		wantErr   bool
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
			name:      "both flags set",
			allRes:    true,
			resources: []string{"pods"},
			wantErr:   false,
		},
		{
			name:      "neither flag set - error",
			allRes:    false,
			resources: []string{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExportFlags(tt.allRes, tt.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExportFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildResourceMap(t *testing.T) {
	resources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment"},
		{Name: "services", Group: "", Version: "v1", Kind: "Service"},
	}

	resourceMap := buildResourceMap(resources)

	if len(resourceMap) != len(resources) {
		t.Errorf("buildResourceMap() returned %d entries, want %d", len(resourceMap), len(resources))
	}

	// Test that all resources are in the map
	for _, res := range resources {
		if _, found := resourceMap[res.Name]; !found {
			t.Errorf("Resource %s not found in map", res.Name)
		}
	}

	// Test specific resource lookup
	if pod, found := resourceMap["pods"]; !found {
		t.Error("pods not found in map")
	} else if pod.Kind != "Pod" {
		t.Errorf("pods Kind = %s, want Pod", pod.Kind)
	}
}

func TestSelectRequestedResources(t *testing.T) {
	resourceMap := map[string]k8s.ResourceInfo{
		"pods":        {Name: "pods", Kind: "Pod"},
		"deployments": {Name: "deployments", Kind: "Deployment"},
		"services":    {Name: "services", Kind: "Service"},
	}

	tests := []struct {
		name            string
		requested       []string
		wantSelectedLen int
		wantNotFoundLen int
		wantNotFound    []string
	}{
		{
			name:            "all found",
			requested:       []string{"pods", "services"},
			wantSelectedLen: 2,
			wantNotFoundLen: 0,
			wantNotFound:    []string{},
		},
		{
			name:            "some not found",
			requested:       []string{"pods", "invalid", "services"},
			wantSelectedLen: 2,
			wantNotFoundLen: 1,
			wantNotFound:    []string{"invalid"},
		},
		{
			name:            "none found",
			requested:       []string{"invalid1", "invalid2"},
			wantSelectedLen: 0,
			wantNotFoundLen: 2,
			wantNotFound:    []string{"invalid1", "invalid2"},
		},
		{
			name:            "empty requested",
			requested:       []string{},
			wantSelectedLen: 0,
			wantNotFoundLen: 0,
			wantNotFound:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selected, notFound := selectRequestedResources(resourceMap, tt.requested)

			if len(selected) != tt.wantSelectedLen {
				t.Errorf("selectRequestedResources() selected %d resources, want %d", len(selected), tt.wantSelectedLen)
			}

			if len(notFound) != tt.wantNotFoundLen {
				t.Errorf("selectRequestedResources() notFound %d resources, want %d", len(notFound), tt.wantNotFoundLen)
			}

			// Verify notFound contents
			for i, nf := range notFound {
				if i < len(tt.wantNotFound) && nf != tt.wantNotFound[i] {
					t.Errorf("notFound[%d] = %s, want %s", i, nf, tt.wantNotFound[i])
				}
			}
		})
	}
}

func TestShouldProcessResource(t *testing.T) {
	tests := []struct {
		name      string
		resource  k8s.ResourceInfo
		namespace string
		want      bool
	}{
		{
			name:      "namespaced resource with namespace",
			resource:  k8s.ResourceInfo{Name: "pods", Namespaced: true},
			namespace: "default",
			want:      true,
		},
		{
			name:      "cluster-scoped resource with namespace",
			resource:  k8s.ResourceInfo{Name: "nodes", Namespaced: false},
			namespace: "default",
			want:      false,
		},
		{
			name:      "cluster-scoped resource without namespace",
			resource:  k8s.ResourceInfo{Name: "nodes", Namespaced: false},
			namespace: "",
			want:      true,
		},
		{
			name:      "namespaced resource without namespace",
			resource:  k8s.ResourceInfo{Name: "pods", Namespaced: true},
			namespace: "",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldProcessResource(tt.resource, tt.namespace)
			if got != tt.want {
				t.Errorf("shouldProcessResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatOutputMessage(t *testing.T) {
	tests := []struct {
		name         string
		dryRun       bool
		namespace    string
		resourceType string
		resourceName string
		wantContains string
	}{
		{
			name:         "dry-run mode",
			dryRun:       true,
			namespace:    "default",
			resourceType: "pods",
			resourceName: "test-pod",
			wantContains: "[DRY-RUN]",
		},
		{
			name:         "normal mode",
			dryRun:       false,
			namespace:    "default",
			resourceType: "deployments",
			resourceName: "test-deployment",
			wantContains: "Exported:",
		},
		{
			name:         "different namespace",
			dryRun:       false,
			namespace:    "kube-system",
			resourceType: "services",
			resourceName: "kube-dns",
			wantContains: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOutputMessage(tt.dryRun, tt.namespace, tt.resourceType, tt.resourceName)

			if !containsSubstr(got, tt.wantContains) {
				t.Errorf("formatOutputMessage() = %s, want to contain %s", got, tt.wantContains)
			}

			// Verify format structure
			if !containsSubstr(got, tt.namespace) {
				t.Errorf("formatOutputMessage() missing namespace %s", tt.namespace)
			}
			if !containsSubstr(got, tt.resourceType) {
				t.Errorf("formatOutputMessage() missing resourceType %s", tt.resourceType)
			}
			if !containsSubstr(got, tt.resourceName) {
				t.Errorf("formatOutputMessage() missing resourceName %s", tt.resourceName)
			}
		})
	}
}

// Benchmark tests
func BenchmarkBuildResourceMap(b *testing.B) {
	resources := make([]k8s.ResourceInfo, 100)
	for i := 0; i < 100; i++ {
		resources[i] = k8s.ResourceInfo{
			Name:    string(rune('a' + i%26)),
			Version: "v1",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildResourceMap(resources)
	}
}

func BenchmarkSelectRequestedResources(b *testing.B) {
	resourceMap := make(map[string]k8s.ResourceInfo)
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		resourceMap[name] = k8s.ResourceInfo{Name: name}
	}

	requested := []string{"a", "b", "c", "d", "e", "invalid"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selectRequestedResources(resourceMap, requested)
	}
}

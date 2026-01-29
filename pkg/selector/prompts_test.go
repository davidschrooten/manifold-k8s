package selector

import (
	"testing"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
)

func TestFormatContextOptions(t *testing.T) {
	contexts := []string{"context1", "context2", "context3"}
	current := "context2"

	options := FormatContextOptions(contexts, current)

	if len(options) != len(contexts) {
		t.Errorf("FormatContextOptions() returned %d options, want %d", len(options), len(contexts))
	}

	// Check that current context is marked
	found := false
	for _, opt := range options {
		if opt == "context2 (current)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("FormatContextOptions() did not mark current context")
	}

	// Check that other contexts are not marked
	for _, opt := range options {
		if opt == "context1" || opt == "context3" {
			continue
		}
		if opt != "context2 (current)" {
			t.Errorf("FormatContextOptions() returned unexpected option: %s", opt)
		}
	}
}

func TestFormatResourceOptions(t *testing.T) {
	resources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod", Namespaced: true},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment", Namespaced: true},
		{Name: "customresources", Group: "custom.io", Version: "v1alpha1", Kind: "CustomResource", Namespaced: true},
	}

	options := FormatResourceOptions(resources)

	if len(options) != len(resources) {
		t.Errorf("FormatResourceOptions() returned %d options, want %d", len(options), len(resources))
	}

	expectedOptions := map[string]bool{
		"pods (v1/Pod)":                                  true,
		"deployments (apps/v1/Deployment)":               true,
		"customresources (custom.io/v1alpha1/CustomResource)": true,
	}

	for _, opt := range options {
		if !expectedOptions[opt] {
			t.Errorf("FormatResourceOptions() returned unexpected option: %s", opt)
		}
	}
}

func TestParseSelectedContext(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
	}{
		{
			name:  "parse current context",
			input: "context1 (current)",
			want:  "context1",
		},
		{
			name:  "parse regular context",
			input: "context2",
			want:  "context2",
		},
		{
			name:  "parse context with spaces",
			input: "my-context (current)",
			want:  "my-context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSelectedContext(tt.input)
			if got != tt.want {
				t.Errorf("ParseSelectedContext(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSelectedResource(t *testing.T) {
	resources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment"},
	}

	tests := []struct {
		name     string
		input    string
		want     *k8s.ResourceInfo
		wantNil  bool
	}{
		{
			name:  "parse core resource",
			input: "pods (v1/Pod)",
			want:  &k8s.ResourceInfo{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
		},
		{
			name:  "parse apps resource",
			input: "deployments (apps/v1/Deployment)",
			want:  &k8s.ResourceInfo{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment"},
		},
		{
			name:    "parse unknown resource",
			input:   "unknown (v1/Unknown)",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSelectedResource(tt.input, resources)
			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseSelectedResource(%s) = %v, want nil", tt.input, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ParseSelectedResource(%s) = nil, want %v", tt.input, tt.want)
			}
			if got.Name != tt.want.Name || got.Group != tt.want.Group || got.Version != tt.want.Version || got.Kind != tt.want.Kind {
				t.Errorf("ParseSelectedResource(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty directory",
			input:   "",
			wantErr: true,
		},
		{
			name:    "valid directory path",
			input:   "/tmp/test",
			wantErr: false,
		},
		{
			name:    "relative directory path",
			input:   "./output",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectory(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectory(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}

	// Test invalid input type
	t.Run("invalid input type", func(t *testing.T) {
		err := ValidateDirectory(123)
		if err == nil {
			t.Error("ValidateDirectory() expected error for invalid type, got nil")
		}
	})
}

func TestPromptContextSelection_EmptyContexts(t *testing.T) {
	// Test error case with no contexts
	_, err := PromptContextSelection([]string{}, "")
	if err == nil {
		t.Error("PromptContextSelection() with empty contexts should return error")
	}
}

func TestPromptNamespaceSelection_EmptyNamespaces(t *testing.T) {
	// Test error case with no namespaces
	_, err := PromptNamespaceSelection([]string{})
	if err == nil {
		t.Error("PromptNamespaceSelection() with empty namespaces should return error")
	}
}

func TestPromptResourceSelection_EmptyResources(t *testing.T) {
	// Test error case with no resources
	_, err := PromptResourceSelection([]k8s.ResourceInfo{})
	if err == nil {
		t.Error("PromptResourceSelection() with empty resources should return error")
	}
}

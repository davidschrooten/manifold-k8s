package selector

import (
	"errors"
	"testing"

	"github.com/AlecAivazis/survey/v2"
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
		"pods (v1/Pod)":                                       true,
		"deployments (apps/v1/Deployment)":                    true,
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
		name  string
		input string
		want  string
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
		name    string
		input   string
		want    *k8s.ResourceInfo
		wantNil bool
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

// Mock prompter for testing
type mockPrompter struct {
	contexts   []string
	namespaces []string
	resources  []k8s.ResourceInfo
	directory  string
	confirmed  bool
	err        error
}

func (m *mockPrompter) PromptContextSelection(contexts []string, currentContext string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.contexts, nil
}

func (m *mockPrompter) PromptNamespaceSelection(namespaces []string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.namespaces, nil
}

func (m *mockPrompter) PromptResourceSelection(resources []k8s.ResourceInfo) ([]k8s.ResourceInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resources, nil
}

func (m *mockPrompter) PromptDirectorySelection(defaultDir string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.directory, nil
}

func (m *mockPrompter) PromptConfirmation(message string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.confirmed, nil
}

func TestDefaultPrompter(t *testing.T) {
	prompter := NewDefaultPrompter()
	if prompter == nil {
		t.Fatal("NewDefaultPrompter() returned nil")
	}
}

func TestMockPrompter_ContextSelection(t *testing.T) {
	mock := &mockPrompter{
		contexts: []string{"context1", "context2"},
	}

	result, err := mock.PromptContextSelection([]string{"ctx1", "ctx2"}, "ctx1")
	if err != nil {
		t.Errorf("PromptContextSelection() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("PromptContextSelection() returned %d contexts, want 2", len(result))
	}
}

func TestMockPrompter_NamespaceSelection(t *testing.T) {
	mock := &mockPrompter{
		namespaces: []string{"default", "kube-system"},
	}

	result, err := mock.PromptNamespaceSelection([]string{"default", "kube-system"})
	if err != nil {
		t.Errorf("PromptNamespaceSelection() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("PromptNamespaceSelection() returned %d namespaces, want 2", len(result))
	}
}

func TestMockPrompter_ResourceSelection(t *testing.T) {
	resources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
	}

	mock := &mockPrompter{
		resources: resources,
	}

	result, err := mock.PromptResourceSelection(resources)
	if err != nil {
		t.Errorf("PromptResourceSelection() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("PromptResourceSelection() returned %d resources, want 1", len(result))
	}
}

func TestMockPrompter_DirectorySelection(t *testing.T) {
	mock := &mockPrompter{
		directory: "/tmp/output",
	}

	result, err := mock.PromptDirectorySelection("./output")
	if err != nil {
		t.Errorf("PromptDirectorySelection() error = %v", err)
	}
	if result != "/tmp/output" {
		t.Errorf("PromptDirectorySelection() = %s, want /tmp/output", result)
	}
}

func TestMockPrompter_Confirmation(t *testing.T) {
	tests := []struct {
		name      string
		confirmed bool
	}{
		{"confirmed", true},
		{"not confirmed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPrompter{
				confirmed: tt.confirmed,
			}

			result, err := mock.PromptConfirmation("Proceed?")
			if err != nil {
				t.Errorf("PromptConfirmation() error = %v", err)
			}
			if result != tt.confirmed {
				t.Errorf("PromptConfirmation() = %v, want %v", result, tt.confirmed)
			}
		})
	}
}

func TestPromptResourceSelection_NoSelection(t *testing.T) {
	// Test when user selects nothing and survey returns empty selection
	resources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
	}

	// We can't easily mock survey.AskOne, but we can test the error path
	// when the function returns no selections
	// This would require refactoring the actual implementation to be testable
	// For now, we document that this case is covered by integration tests
	_ = resources
}

func TestPromptNamespaceSelection_NoSelection(t *testing.T) {
	// Similar to above - documents that no-selection case exists
	// but requires integration testing or survey mocking
}

// Test with mocked survey.AskOne
func TestPromptContextSelection_WithMockedAsker(t *testing.T) {
	tests := []struct {
		name            string
		contexts        []string
		currentContext  string
		mockedResponse  []string
		mockedError     error
		wantErr         bool
		wantResultCount int
	}{
		{
			name:            "successful selection",
			contexts:        []string{"ctx1", "ctx2"},
			currentContext:  "ctx1",
			mockedResponse:  []string{"ctx1 (current)", "ctx2"},
			mockedError:     nil,
			wantErr:         false,
			wantResultCount: 2,
		},
		{
			name:           "survey error",
			contexts:       []string{"ctx1"},
			currentContext: "ctx1",
			mockedError:    errors.New("user cancelled"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock asker
			mockAsker := func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				if tt.mockedError != nil {
					return tt.mockedError
				}
				// Simulate user selection
				if selected, ok := response.(*[]string); ok {
					*selected = tt.mockedResponse
				}
				return nil
			}

			result, err := promptContextSelectionWithAsker(mockAsker, tt.contexts, tt.currentContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptContextSelectionWithAsker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantResultCount {
				t.Errorf("promptContextSelectionWithAsker() returned %d contexts, want %d", len(result), tt.wantResultCount)
			}
		})
	}
}

func TestPromptNamespaceSelection_WithMockedAsker(t *testing.T) {
	tests := []struct {
		name            string
		namespaces      []string
		mockedResponse  []string
		mockedError     error
		wantErr         bool
		wantResultCount int
	}{
		{
			name:            "successful selection",
			namespaces:      []string{"default", "kube-system"},
			mockedResponse:  []string{"default"},
			mockedError:     nil,
			wantErr:         false,
			wantResultCount: 1,
		},
		{
			name:        "survey error",
			namespaces:  []string{"default"},
			mockedError: errors.New("interrupted"),
			wantErr:     true,
		},
		{
			name:           "no selection",
			namespaces:     []string{"default"},
			mockedResponse: []string{},
			mockedError:    nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAsker := func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				if tt.mockedError != nil {
					return tt.mockedError
				}
				if selected, ok := response.(*[]string); ok {
					*selected = tt.mockedResponse
				}
				return nil
			}

			result, err := promptNamespaceSelectionWithAsker(mockAsker, tt.namespaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptNamespaceSelectionWithAsker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantResultCount {
				t.Errorf("promptNamespaceSelectionWithAsker() returned %d namespaces, want %d", len(result), tt.wantResultCount)
			}
		})
	}
}

func TestPromptResourceSelection_WithMockedAsker(t *testing.T) {
	resources := []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod"},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment"},
	}

	tests := []struct {
		name            string
		mockedResponse  []string
		mockedError     error
		wantErr         bool
		wantResultCount int
	}{
		{
			name:            "successful selection",
			mockedResponse:  []string{"pods (v1/Pod)"},
			mockedError:     nil,
			wantErr:         false,
			wantResultCount: 1,
		},
		{
			name:        "survey error",
			mockedError: errors.New("cancelled"),
			wantErr:     true,
		},
		{
			name:           "no selection",
			mockedResponse: []string{},
			mockedError:    nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAsker := func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				if tt.mockedError != nil {
					return tt.mockedError
				}
				if selected, ok := response.(*[]string); ok {
					*selected = tt.mockedResponse
				}
				return nil
			}

			result, err := promptResourceSelectionWithAsker(mockAsker, resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptResourceSelectionWithAsker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantResultCount {
				t.Errorf("promptResourceSelectionWithAsker() returned %d resources, want %d", len(result), tt.wantResultCount)
			}
		})
	}
}

func TestPromptDirectorySelection_WithMockedAsker(t *testing.T) {
	tests := []struct {
		name           string
		defaultDir     string
		mockedResponse string
		mockedError    error
		wantErr        bool
		wantResult     string
	}{
		{
			name:           "successful selection",
			defaultDir:     "./output",
			mockedResponse: "/tmp/manifests",
			mockedError:    nil,
			wantErr:        false,
			wantResult:     "/tmp/manifests",
		},
		{
			name:        "survey error",
			defaultDir:  "./output",
			mockedError: errors.New("interrupted"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAsker := func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				if tt.mockedError != nil {
					return tt.mockedError
				}
				if dir, ok := response.(*string); ok {
					*dir = tt.mockedResponse
				}
				return nil
			}

			result, err := promptDirectorySelectionWithAsker(mockAsker, tt.defaultDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptDirectorySelectionWithAsker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantResult {
				t.Errorf("promptDirectorySelectionWithAsker() = %s, want %s", result, tt.wantResult)
			}
		})
	}
}

func TestPromptConfirmation_WithMockedAsker(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		mockedResponse bool
		mockedError    error
		wantErr        bool
		wantResult     bool
	}{
		{
			name:           "confirmed",
			message:        "Proceed?",
			mockedResponse: true,
			mockedError:    nil,
			wantErr:        false,
			wantResult:     true,
		},
		{
			name:           "not confirmed",
			message:        "Proceed?",
			mockedResponse: false,
			mockedError:    nil,
			wantErr:        false,
			wantResult:     false,
		},
		{
			name:        "survey error",
			message:     "Proceed?",
			mockedError: errors.New("cancelled"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAsker := func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				if tt.mockedError != nil {
					return tt.mockedError
				}
				if confirmed, ok := response.(*bool); ok {
					*confirmed = tt.mockedResponse
				}
				return nil
			}

			result, err := promptConfirmationWithAsker(mockAsker, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptConfirmationWithAsker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantResult {
				t.Errorf("promptConfirmationWithAsker() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

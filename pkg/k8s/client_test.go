package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKubeConfig(t *testing.T) {
	tests := []struct {
		name        string
		kubeconfigPath string
		setup       func(t *testing.T) string
		wantErr     bool
	}{
		{
			name: "load default kubeconfig",
			setup: func(t *testing.T) string {
				// Create a temporary kubeconfig
				tmpDir := t.TempDir()
				kubeconfigPath := filepath.Join(tmpDir, "config")
				content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
				if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				return kubeconfigPath
			},
			wantErr: false,
		},
		{
			name: "load from home directory when path is empty",
			setup: func(t *testing.T) string {
				// Create kubeconfig in home/.kube/config
				home, err := os.UserHomeDir()
				if err != nil {
					t.Skip("Cannot get home directory")
				}
				kubeDir := filepath.Join(home, ".kube")
				if err := os.MkdirAll(kubeDir, 0755); err != nil {
					t.Fatal(err)
				}
				kubeconfigPath := filepath.Join(kubeDir, "config")
				// Save existing config if it exists
				existingConfig, _ := os.ReadFile(kubeconfigPath)
				t.Cleanup(func() {
					if len(existingConfig) > 0 {
						os.WriteFile(kubeconfigPath, existingConfig, 0600)
					} else {
						os.Remove(kubeconfigPath)
					}
				})
				content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
				if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				return "" // Empty path to trigger default behavior
			},
			wantErr: false,
		},
		{
			name:        "fail on non-existent kubeconfig",
			kubeconfigPath: "/non/existent/path",
			setup:       func(t *testing.T) string { return "/non/existent/path" },
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := tt.setup(t)
			_, err := LoadKubeConfig(kubeconfigPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadKubeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetContexts(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://cluster1:6443
  name: cluster1
- cluster:
    server: https://cluster2:6443
  name: cluster2
contexts:
- context:
    cluster: cluster1
    user: user1
  name: context1
- context:
    cluster: cluster2
    user: user2
  name: context2
current-context: context1
users:
- name: user1
  user:
    token: token1
- name: user2
  user:
    token: token2
`
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadKubeConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("LoadKubeConfig() error = %v", err)
	}

	contexts := GetContexts(config)
	if len(contexts) != 2 {
		t.Errorf("GetContexts() returned %d contexts, want 2", len(contexts))
	}

	expectedContexts := map[string]bool{"context1": true, "context2": true}
	for _, ctx := range contexts {
		if !expectedContexts[ctx] {
			t.Errorf("GetContexts() returned unexpected context: %s", ctx)
		}
	}
}

func TestGetCurrentContext(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadKubeConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("LoadKubeConfig() error = %v", err)
	}

	currentContext := GetCurrentContext(config)
	if currentContext != "test-context" {
		t.Errorf("GetCurrentContext() = %v, want %v", currentContext, "test-context")
	}
}

func TestNewClient(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadKubeConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("LoadKubeConfig() error = %v", err)
	}

	client, err := NewClient(config, "test-context")
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
}

func TestNewClient_InvalidContext(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadKubeConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("LoadKubeConfig() error = %v", err)
	}

	_, err = NewClient(config, "non-existent-context")
	if err == nil {
		t.Error("NewClient() expected error for invalid context, got nil")
	}
}

func TestNewClient_InvalidClusterConfig(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	// Create invalid config with missing cluster
	content := `apiVersion: v1
kind: Config
clusters: []
contexts:
- context:
    cluster: missing-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadKubeConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("LoadKubeConfig() error = %v", err)
	}

	_, err = NewClient(config, "test-context")
	if err == nil {
		t.Error("NewClient() expected error for invalid cluster config, got nil")
	}
}

func TestNewClient_InvalidServerURL(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	// Create config with invalid server URL
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: ":::invalid-url"
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadKubeConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("LoadKubeConfig() error = %v", err)
	}

	_, err = NewClient(config, "test-context")
	if err == nil {
		t.Error("NewClient() expected error for invalid server URL, got nil")
	}
}

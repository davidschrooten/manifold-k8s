package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Client wraps Kubernetes clients
type Client struct {
	Clientset     kubernetes.Interface
	DynamicClient dynamic.Interface
	RESTConfig    *rest.Config
	Context       string
}

// LoadKubeConfig loads kubeconfig from the specified path
// If path is empty, uses default location ($HOME/.kube/config)
func LoadKubeConfig(kubeconfigPath string) (*api.Config, error) {
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
	}

	return config, nil
}

// GetContexts returns a list of all available contexts from kubeconfig
func GetContexts(config *api.Config) []string {
	contexts := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}

// GetCurrentContext returns the current context from kubeconfig
func GetCurrentContext(config *api.Config) string {
	return config.CurrentContext
}

// NewClient creates a new Kubernetes client for the specified context
func NewClient(config *api.Config, context string) (*Client, error) {
	// Validate that the context exists
	if _, exists := config.Contexts[context]; !exists {
		return nil, fmt.Errorf("context %s not found in kubeconfig", context)
	}

	// Create a client config for the specific context
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{
		CurrentContext: context,
	})

	// Get the REST config
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config for context %s: %w", context, err)
	}

	// Create the standard clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create the dynamic client for custom resources
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		Clientset:     clientset,
		DynamicClient: dynamicClient,
		RESTConfig:    restConfig,
		Context:       context,
	}, nil
}

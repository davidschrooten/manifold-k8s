package cmd

import (
	"context"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
)

// mockKubeConfig returns a mock kubeconfig
func mockKubeConfig() *api.Config {
	config := api.NewConfig()
	config.Clusters["test-cluster"] = &api.Cluster{
		Server: "https://localhost:6443",
	}
	config.Contexts["test-context"] = &api.Context{
		Cluster:  "test-cluster",
		AuthInfo: "test-user",
	}
	config.AuthInfos["test-user"] = &api.AuthInfo{
		Token: "test-token",
	}
	config.CurrentContext = "test-context"
	return config
}

// mockK8sClient returns a mock k8s client
func mockK8sClient() *k8s.Client {
	return &k8s.Client{
		Clientset:     &kubernetes.Clientset{},
		DynamicClient: &mockDynamicClient{},
		RESTConfig:    nil,
	}
}

// mockDynamicClient is a stub dynamic client
type mockDynamicClient struct{}

func (m *mockDynamicClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &mockNamespaceableResource{gvr: resource}
}

type mockNamespaceableResource struct {
	gvr       schema.GroupVersionResource
	namespace string
}

func (m *mockNamespaceableResource) Namespace(ns string) dynamic.ResourceInterface {
	return &mockNamespaceableResource{gvr: m.gvr, namespace: ns}
}

func (m *mockNamespaceableResource) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (m *mockNamespaceableResource) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m *mockNamespaceableResource) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	// Return mock data based on resource type
	items := []unstructured.Unstructured{}

	switch m.gvr.Resource {
	case "pods":
		items = append(items, unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name":      "test-pod-1",
					"namespace": m.namespace,
				},
			},
		})
	case "deployments":
		items = append(items, unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "test-deployment-1",
					"namespace": m.namespace,
				},
			},
		})
	}

	return &unstructured.UnstructuredList{Items: items}, nil
}

func (m *mockNamespaceableResource) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResource) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

// mockDiscoveredResources returns mock discovered resources
func mockDiscoveredResources() []k8s.ResourceInfo {
	return []k8s.ResourceInfo{
		{Name: "pods", Group: "", Version: "v1", Kind: "Pod", Namespaced: true},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment", Namespaced: true},
		{Name: "services", Group: "", Version: "v1", Kind: "Service", Namespaced: true},
	}
}

// mockNamespaces returns mock namespaces
func mockNamespaces() []string {
	return []string{"default", "kube-system"}
}

// enableStubs enables all stubs for testing
func enableStubs() {
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return mockKubeConfig(), nil
	}

	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return mockK8sClient(), nil
	}

	stubDiscoverResources = func(discovery.DiscoveryInterface) ([]k8s.ResourceInfo, error) {
		return mockDiscoveredResources(), nil
	}

	stubGetNamespaces = func(ctx context.Context, client *k8s.Client) ([]string, error) {
		return mockNamespaces(), nil
	}
}

// disableStubs disables all stubs
func disableStubs() {
	stubLoadKubeConfig = nil
	stubNewClient = nil
	stubDiscoverResources = nil
	stubGetNamespaces = nil
}

package cmd

import (
	"context"

	"github.com/davidschrooten/manifold-k8s/pkg/helm"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Stub functions that can be replaced for testing
// These are set to nil by default and only populated in tests
var (
	stubLoadKubeConfig    func(string) (*api.Config, error)
	stubNewClient         func(*api.Config, string) (*k8s.Client, error)
	stubDiscoverResources func(discovery.DiscoveryInterface) ([]k8s.ResourceInfo, error)
	stubGetNamespaces     func(context.Context, *k8s.Client) ([]string, error)
	stubListHelmReleases  func(namespace string) ([]helm.Release, error)
	stubGetHelmValues     func(releaseName, namespace string) (string, error)
)

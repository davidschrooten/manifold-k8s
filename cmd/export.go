package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/davidschrooten/manifold-k8s/pkg/exporter"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	exportDryRun     bool
	exportOutputDir  string
	exportCtx        string
	exportNamespaces []string
	exportResources  []string
	exportAllRes     bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export Kubernetes manifests non-interactively",
	Long: `Export Kubernetes manifests by specifying context, namespaces, and resources.

This command requires all parameters to be provided via flags and does not prompt for input.
It is designed for scripting and CI/CD pipelines.

Examples:
  manifold-k8s export --context prod --namespaces default,kube-system --resources pods,deployments -o ./output
  manifold-k8s export --context staging --namespaces myapp --all-resources -o ./backup`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVar(&exportDryRun, "dry-run", false, "preview what would be exported without writing files")
	exportCmd.Flags().StringVarP(&exportOutputDir, "output", "o", "", "output directory (required)")
	exportCmd.Flags().StringVarP(&exportCtx, "context", "c", "", "kubernetes context (required)")
	exportCmd.Flags().StringSliceVarP(&exportNamespaces, "namespaces", "n", nil, "namespaces to export (comma-separated, required)")
	exportCmd.Flags().StringSliceVarP(&exportResources, "resources", "r", nil, "resource types to export (comma-separated, e.g. pods,deployments)")
	exportCmd.Flags().BoolVarP(&exportAllRes, "all-resources", "a", false, "export all resource types")

	_ = exportCmd.MarkFlagRequired("context")
	_ = exportCmd.MarkFlagRequired("namespaces")
	_ = exportCmd.MarkFlagRequired("output")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	// Validate required flags
	if err := validateExportFlags(exportAllRes, exportResources); err != nil {
		return err
	}

	// Load kubeconfig (use stub if available)
	kubeconfigPath := viper.GetString("kubeconfig")
	var err error
	var config *api.Config
	if stubLoadKubeConfig != nil {
		config, err = stubLoadKubeConfig(kubeconfigPath)
	} else {
		config, err = k8s.LoadKubeConfig(kubeconfigPath)
	}
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create client for specified context (use stub if available)
	var client *k8s.Client
	if stubNewClient != nil {
		client, err = stubNewClient(config, exportCtx)
	} else {
		client, err = k8s.NewClient(config, exportCtx)
	}
	if err != nil {
		return fmt.Errorf("failed to create client for context %s: %w", exportCtx, err)
	}

	fmt.Printf("Using context: %s\n", exportCtx)

	// Discover resources (use stub if available)
	var discoveredResources []k8s.ResourceInfo
	if stubDiscoverResources != nil {
		discoveredResources, err = stubDiscoverResources(client.Clientset.Discovery())
	} else {
		discoveredResources, err = k8s.DiscoverResources(client.Clientset.Discovery())
	}
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	// Filter resources based on flags
	var selectedResources []k8s.ResourceInfo
	if exportAllRes {
		selectedResources = discoveredResources
		fmt.Printf("Exporting all resource types (%d types)\n", len(selectedResources))
	} else {
		// Build map and select requested resources
		resourceMap := buildResourceMap(discoveredResources)
		var notFound []string
		selectedResources, notFound = selectRequestedResources(resourceMap, exportResources)

		// Warn about not found resources
		for _, resName := range notFound {
			fmt.Fprintf(os.Stderr, "Warning: resource type %s not found in cluster\n", resName)
		}

		if len(selectedResources) == 0 {
			return fmt.Errorf("no valid resource types found")
		}
		fmt.Printf("Exporting %d resource type(s): %v\n", len(selectedResources), exportResources)
	}

	fmt.Printf("Exporting from %d namespace(s): %v\n", len(exportNamespaces), exportNamespaces)

	// Create exporter
	exp := exporter.NewExporter(exportOutputDir)

	// Fetch and export resources
	fmt.Println("\nExporting manifests...")
	for _, namespace := range exportNamespaces {
		for _, resource := range selectedResources {
			if !shouldProcessResource(resource, namespace) {
				continue
			}

			gvr := resource.GroupVersionResource()

			// List resources
			var resourceList *unstructured.UnstructuredList
			if resource.Namespaced {
				resourceList, err = client.DynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
			} else {
				resourceList, err = client.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to list %s in %s: %v\n", resource.Name, namespace, err)
				continue
			}

			// Export each resource
			for _, item := range resourceList.Items {
				if exportDryRun {
					fmt.Println(formatOutputMessage(true, namespace, resource.Name, item.GetName()))
					continue
				}

				if err := exp.ExportResource(ctx, &item, gvr, namespace); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to export %s/%s: %v\n", resource.Name, item.GetName(), err)
					continue
				}
				fmt.Println(formatOutputMessage(false, namespace, resource.Name, item.GetName()))
			}
		}
	}

	// Print summary
	if !exportDryRun {
		fmt.Printf("\n%s\n", exp.Summary())
	}

	return nil
}

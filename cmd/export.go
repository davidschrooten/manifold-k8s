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

	exportCmd.MarkFlagRequired("context")
	exportCmd.MarkFlagRequired("namespaces")
	exportCmd.MarkFlagRequired("output")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	// Validate required flags
	if !exportAllRes && len(exportResources) == 0 {
		return fmt.Errorf("either --resources or --all-resources is required")
	}

	// Load kubeconfig
	kubeconfigPath := viper.GetString("kubeconfig")
	config, err := k8s.LoadKubeConfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create client for specified context
	client, err := k8s.NewClient(config, exportCtx)
	if err != nil {
		return fmt.Errorf("failed to create client for context %s: %w", exportCtx, err)
	}

	fmt.Printf("Using context: %s\n", exportCtx)

	// Discover resources
	discoveredResources, err := k8s.DiscoverResources(client.Clientset.Discovery())
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	// Filter resources based on flags
	var selectedResources []k8s.ResourceInfo
	if exportAllRes {
		selectedResources = discoveredResources
		fmt.Printf("Exporting all resource types (%d types)\n", len(selectedResources))
	} else {
		// Build map of resource names
		resourceMap := make(map[string]k8s.ResourceInfo)
		for _, res := range discoveredResources {
			resourceMap[res.Name] = res
		}

		// Select requested resources
		for _, resName := range exportResources {
			if res, found := resourceMap[resName]; found {
				selectedResources = append(selectedResources, res)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: resource type %s not found in cluster\n", resName)
			}
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
			if !resource.Namespaced && namespace != "" {
				continue // Skip cluster-scoped resources when processing namespaces
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
					fmt.Printf("[DRY-RUN] Would export: %s/%s/%s\n", namespace, resource.Name, item.GetName())
					continue
				}

				if err := exp.ExportResource(ctx, &item, gvr, namespace); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to export %s/%s: %v\n", resource.Name, item.GetName(), err)
					continue
				}
				fmt.Printf("Exported: %s/%s/%s\n", namespace, resource.Name, item.GetName())
			}
		}
	}

	// Print summary
	if !exportDryRun {
		fmt.Printf("\n%s\n", exp.Summary())
	}

	return nil
}

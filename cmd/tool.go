package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/davidschrooten/manifold-k8s/pkg/exporter"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/davidschrooten/manifold-k8s/pkg/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	dryRun       bool
	outputDir    string
	ctxName      string
	namespaces   []string
	resources    []string
	allResources bool
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download Kubernetes manifests",
	Long: `Download Kubernetes manifests interactively or non-interactively.

Interactive mode (default):
  Select clusters, namespaces, and resource types through prompts.

Non-interactive mode:
  Provide --context, --namespaces, and optionally --resources or --all-resources.
  Example: manifold-k8s download --context prod --namespaces default,kube-system --resources pods,deployments -o ./output`,
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would be downloaded without writing files")
	downloadCmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory (will be prompted if not provided)")
	downloadCmd.Flags().StringVarP(&ctxName, "context", "c", "", "kubernetes context (non-interactive mode)")
	downloadCmd.Flags().StringSliceVarP(&namespaces, "namespaces", "n", nil, "namespaces to export (comma-separated, non-interactive mode)")
	downloadCmd.Flags().StringSliceVarP(&resources, "resources", "r", nil, "resource types to export (comma-separated, e.g. pods,deployments)")
	downloadCmd.Flags().BoolVarP(&allResources, "all-resources", "a", false, "export all resource types (non-interactive mode)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Check if non-interactive mode
	nonInteractive := ctxName != "" || len(namespaces) > 0 || len(resources) > 0 || allResources

	if nonInteractive {
		return runNonInteractive(ctx)
	}

	return runInteractive(ctx)
}

func runNonInteractive(ctx context.Context) error {
	// Validate required flags
	if ctxName == "" {
		return fmt.Errorf("--context is required for non-interactive mode")
	}
	if len(namespaces) == 0 {
		return fmt.Errorf("--namespaces is required for non-interactive mode")
	}
	if !allResources && len(resources) == 0 {
		return fmt.Errorf("either --resources or --all-resources is required for non-interactive mode")
	}
	if outputDir == "" {
		return fmt.Errorf("--output is required for non-interactive mode")
	}

	// Load kubeconfig
	kubeconfigPath := viper.GetString("kubeconfig")
	config, err := k8s.LoadKubeConfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create client for specified context
	client, err := k8s.NewClient(config, ctxName)
	if err != nil {
		return fmt.Errorf("failed to create client for context %s: %w", ctxName, err)
	}

	fmt.Printf("Using context: %s\n", ctxName)

	// Discover resources
	discoveredResources, err := k8s.DiscoverResources(client.Clientset.Discovery())
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	// Filter resources based on flags
	var selectedResources []k8s.ResourceInfo
	if allResources {
		selectedResources = discoveredResources
		fmt.Printf("Exporting all resource types (%d types)\n", len(selectedResources))
	} else {
		// Build map of resource names
		resourceMap := make(map[string]k8s.ResourceInfo)
		for _, res := range discoveredResources {
			resourceMap[res.Name] = res
		}

		// Select requested resources
		for _, resName := range resources {
			if res, found := resourceMap[resName]; found {
				selectedResources = append(selectedResources, res)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: resource type %s not found in cluster\n", resName)
			}
		}

		if len(selectedResources) == 0 {
			return fmt.Errorf("no valid resource types found")
		}
		fmt.Printf("Exporting %d resource type(s): %v\n", len(selectedResources), resources)
	}

	fmt.Printf("Exporting from %d namespace(s): %v\n", len(namespaces), namespaces)

	// Create exporter
	exp := exporter.NewExporter(outputDir)

	// Fetch and export resources
	fmt.Println("\nExporting manifests...")
	for _, namespace := range namespaces {
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
				if dryRun {
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
	if !dryRun {
		fmt.Printf("\n%s\n", exp.Summary())
	}

	return nil
}

func runInteractive(ctx context.Context) error {
	// 1. Load kubeconfig
	kubeconfigPath := viper.GetString("kubeconfig")
	config, err := k8s.LoadKubeConfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	contexts := k8s.GetContexts(config)
	currentContext := k8s.GetCurrentContext(config)

	// 2. Select cluster context(s)
	fmt.Println("\nSelecting cluster context(s)...")
	selectedContexts, err := selector.PromptContextSelection(contexts, currentContext)
	if err != nil {
		return fmt.Errorf("context selection failed: %w", err)
	}

	// 3. Process each selected context
	for _, contextName := range selectedContexts {
		fmt.Printf("\n=== Processing context: %s ===\n", contextName)

		// Create client for this context
		client, err := k8s.NewClient(config, contextName)
		if err != nil {
			return fmt.Errorf("failed to create client for context %s: %w", contextName, err)
		}

		// Get namespaces
		namespaces, err := k8s.GetNamespaces(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}

		// Select namespace(s)
		fmt.Println("\nSelecting namespace(s)...")
		selectedNamespaces, err := selector.PromptNamespaceSelection(namespaces)
		if err != nil {
			return fmt.Errorf("namespace selection failed: %w", err)
		}

		// Discover resources
		fmt.Println("\nDiscovering available resources...")
		resources, err := k8s.DiscoverResources(client.Clientset.Discovery())
		if err != nil {
			return fmt.Errorf("failed to discover resources: %w", err)
		}

		// Select resource type(s)
		fmt.Println("\nSelecting resource type(s)...")
		selectedResources, err := selector.PromptResourceSelection(resources)
		if err != nil {
			return fmt.Errorf("resource selection failed: %w", err)
		}

		// Get or prompt for output directory
		if outputDir == "" {
			defaultDir := fmt.Sprintf("./manifests-%s", contextName)
			outputDir, err = selector.PromptDirectorySelection(defaultDir)
			if err != nil {
				return fmt.Errorf("directory selection failed: %w", err)
			}
		}

		// Create exporter
		exp := exporter.NewExporter(outputDir)

		// Fetch and export resources
		fmt.Println("\nExporting manifests...")
		for _, namespace := range selectedNamespaces {
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
					if dryRun {
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
		if !dryRun {
			fmt.Printf("\n%s\n", exp.Summary())
		}
	}

	return nil
}

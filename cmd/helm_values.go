package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidschrooten/manifold-k8s/pkg/helm"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/davidschrooten/manifold-k8s/pkg/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	helmValuesDryRun    bool
	helmValuesOutputDir string
)

var helmValuesCmd = &cobra.Command{
	Use:   "helm-values",
	Short: "Interactively export Helm release values",
	Long: `Export Helm release values by interactively selecting:
- Cluster(s) from kubeconfig
- Namespace(s)
- Helm release(s)
- Target directory

This command requires the helm CLI to be installed and available in PATH.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !helm.IsHelmInstalled() {
			return fmt.Errorf("helm CLI is not installed or not in PATH. Please install helm: https://helm.sh/docs/intro/install/")
		}
		return nil
	},
	RunE: runHelmValuesInteractive,
}

func init() {
	rootCmd.AddCommand(helmValuesCmd)

	helmValuesCmd.Flags().BoolVar(&helmValuesDryRun, "dry-run", false, "preview what would be downloaded without writing files")
	helmValuesCmd.Flags().StringVarP(&helmValuesOutputDir, "output", "o", "", "output directory (will be prompted if not provided)")
}

// runHelmValuesInteractive is excluded from coverage as it requires user interaction
// coverage:ignore
func runHelmValuesInteractive(cmd *cobra.Command, args []string) error {
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

	contexts := k8s.GetContexts(config)
	currentContext := k8s.GetCurrentContext(config)

	// Select cluster context(s)
	fmt.Println("\nSelecting cluster context(s)...")
	selectedContexts, err := selector.PromptContextSelection(contexts, currentContext)
	if err != nil {
		return fmt.Errorf("context selection failed: %w", err)
	}

	// Process each selected context
	for _, contextName := range selectedContexts {
		fmt.Printf("\n=== Processing context: %s ===\n", contextName)

		// Create client for this context
		var client *k8s.Client
		if stubNewClient != nil {
			client, err = stubNewClient(config, contextName)
		} else {
			client, err = k8s.NewClient(config, contextName)
		}
		if err != nil {
			return fmt.Errorf("failed to create client for context %s: %w", contextName, err)
		}

		// Get namespaces
		ctx := cmd.Context()
		var namespaces []string
		if stubGetNamespaces != nil {
			namespaces, err = stubGetNamespaces(ctx, client)
		} else {
			namespaces, err = k8s.GetNamespaces(ctx, client)
		}
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}

		// Select namespace(s)
		fmt.Println("\nSelecting namespace(s)...")
		selectedNamespaces, err := selector.PromptNamespaceSelection(namespaces)
		if err != nil {
			return fmt.Errorf("namespace selection failed: %w", err)
		}

		// Get or prompt for output directory
		outputDir := helmValuesOutputDir
		if outputDir == "" {
			defaultDir := fmt.Sprintf("./helm-values-%s", contextName)
			outputDir, err = selector.PromptDirectorySelection(defaultDir)
			if err != nil {
				return fmt.Errorf("directory selection failed: %w", err)
			}
		}

		// Process each namespace
		for _, namespace := range selectedNamespaces {
			fmt.Printf("\n--- Namespace: %s ---\n", namespace)

			// List Helm releases in this namespace
			var releases []helm.Release
			if stubListHelmReleases != nil {
				releases, err = stubListHelmReleases(namespace)
			} else {
				releases, err = helm.ListReleasesWithContext(namespace, contextName)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to list Helm releases in %s: %v\n", namespace, err)
				continue
			}

			if len(releases) == 0 {
				fmt.Printf("No Helm releases found in namespace %s\n", namespace)
				continue
			}

			fmt.Printf("Found %d Helm release(s) in %s\n", len(releases), namespace)

			// Export values for each release
			for _, release := range releases {
				if helmValuesDryRun {
					fmt.Printf("[DRY-RUN] Would export: %s/%s (%s)\n", namespace, release.Name, release.Chart)
					continue
				}

				// Get values
				var values string
				if stubGetHelmValues != nil {
					values, err = stubGetHelmValues(release.Name, namespace)
				} else {
					values, err = helm.GetValuesWithContext(release.Name, namespace, contextName)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to get values for %s: %v\n", release.Name, err)
					continue
				}

				// Write to file
				nsDir := filepath.Join(outputDir, contextName, namespace)
				if err := os.MkdirAll(nsDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create directory %s: %v\n", nsDir, err)
					continue
				}

				filename := filepath.Join(nsDir, fmt.Sprintf("%s-values.yaml", release.Name))
				if err := os.WriteFile(filename, []byte(values), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to write %s: %v\n", filename, err)
					continue
				}

				fmt.Printf("Exported: %s/%s -> %s\n", namespace, release.Name, filename)
			}
		}
	}

	if !helmValuesDryRun {
		fmt.Printf("\n✓ Helm values exported successfully\n")
	}

	return nil
}

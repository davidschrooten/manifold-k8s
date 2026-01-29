package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidschrooten/manifold-k8s/pkg/helm"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	helmExportDryRun     bool
	helmExportOutputDir  string
	helmExportCtx        string
	helmExportNamespaces []string
	helmExportReleases   []string
	helmExportAll        bool
)

var helmValuesExportCmd = &cobra.Command{
	Use:   "helm-values-export",
	Short: "Export Helm release values non-interactively",
	Long: `Export Helm release values by specifying context, namespaces, and releases.

This command requires all parameters to be provided via flags and does not prompt for input.
It is designed for scripting and CI/CD pipelines.

Requires the helm CLI to be installed and available in PATH.

Examples:
  manifold-k8s helm-values-export --context prod --namespaces default --releases myapp -o ./output
  manifold-k8s helm-values-export --context staging --namespaces app1,app2 --all -o ./helm-backup`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !helm.IsHelmInstalled() {
			return fmt.Errorf("helm CLI is not installed or not in PATH. Please install helm: https://helm.sh/docs/intro/install/")
		}
		return nil
	},
	RunE: runHelmValuesExport,
}

func init() {
	rootCmd.AddCommand(helmValuesExportCmd)

	helmValuesExportCmd.Flags().BoolVar(&helmExportDryRun, "dry-run", false, "preview what would be exported without writing files")
	helmValuesExportCmd.Flags().StringVarP(&helmExportOutputDir, "output", "o", "", "output directory (required)")
	helmValuesExportCmd.Flags().StringVarP(&helmExportCtx, "context", "c", "", "kubernetes context (required)")
	helmValuesExportCmd.Flags().StringSliceVarP(&helmExportNamespaces, "namespaces", "n", nil, "namespaces to export (comma-separated, required)")
	helmValuesExportCmd.Flags().StringSliceVarP(&helmExportReleases, "releases", "r", nil, "helm releases to export (comma-separated)")
	helmValuesExportCmd.Flags().BoolVarP(&helmExportAll, "all", "a", false, "export all helm releases")

	_ = helmValuesExportCmd.MarkFlagRequired("context")
	_ = helmValuesExportCmd.MarkFlagRequired("namespaces")
	_ = helmValuesExportCmd.MarkFlagRequired("output")
}

func runHelmValuesExport(cmd *cobra.Command, args []string) error {
	// Validate flags
	if !helmExportAll && len(helmExportReleases) == 0 {
		return fmt.Errorf("must specify either --releases or --all")
	}

	// Load kubeconfig
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

	// Verify context exists
	// Create client to validate context (not used for Helm operations)
	if stubNewClient != nil {
		_, err = stubNewClient(config, helmExportCtx)
	} else {
		_, err = k8s.NewClient(config, helmExportCtx)
	}
	if err != nil {
		return fmt.Errorf("failed to create client for context %s: %w", helmExportCtx, err)
	}

	fmt.Printf("Using context: %s\n", helmExportCtx)
	fmt.Printf("Exporting from %d namespace(s): %v\n", len(helmExportNamespaces), helmExportNamespaces)

	var exportedCount int

	// Process each namespace
	for _, namespace := range helmExportNamespaces {
		fmt.Printf("\n--- Namespace: %s ---\n", namespace)

		// List Helm releases in this namespace
		var releases []helm.Release
		if stubListHelmReleases != nil {
			releases, err = stubListHelmReleases(namespace)
		} else {
			releases, err = helm.ListReleasesWithContext(namespace, helmExportCtx)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to list Helm releases in %s: %v\n", namespace, err)
			continue
		}

		if len(releases) == 0 {
			fmt.Printf("No Helm releases found in namespace %s\n", namespace)
			continue
		}

		// Filter releases if specific ones were requested
		var releasesToExport []helm.Release
		if helmExportAll {
			releasesToExport = releases
			fmt.Printf("Exporting all %d Helm release(s)\n", len(releases))
		} else {
			// Build a map of requested releases
			requestedMap := make(map[string]bool)
			for _, r := range helmExportReleases {
				requestedMap[r] = true
			}

			// Filter
			for _, rel := range releases {
				if requestedMap[rel.Name] {
					releasesToExport = append(releasesToExport, rel)
				}
			}

			if len(releasesToExport) == 0 {
				fmt.Printf("None of the requested releases found in namespace %s\n", namespace)
				continue
			}

			fmt.Printf("Exporting %d Helm release(s): %v\n", len(releasesToExport), helmExportReleases)
		}

		// Export values for each release
		for _, release := range releasesToExport {
			if helmExportDryRun {
				fmt.Printf("[DRY-RUN] Would export: %s/%s (%s)\n", namespace, release.Name, release.Chart)
				continue
			}

			// Get values
			var values string
			if stubGetHelmValues != nil {
				values, err = stubGetHelmValues(release.Name, namespace)
			} else {
				values, err = helm.GetValuesWithContext(release.Name, namespace, helmExportCtx)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to get values for %s: %v\n", release.Name, err)
				continue
			}

			// Write to file
			nsDir := filepath.Join(helmExportOutputDir, helmExportCtx, namespace)
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
			exportedCount++
		}
	}

	if !helmExportDryRun {
		fmt.Printf("\n✓ Exported %d Helm release value(s) to %s\n", exportedCount, helmExportOutputDir)
	}

	return nil
}

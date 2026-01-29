package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	dryRun    bool
	outputDir string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download Kubernetes manifests interactively",
	Long: `Download Kubernetes manifests by interactively selecting:
- Cluster(s) from kubeconfig
- Namespace(s)
- Resource type(s)
- Optionally specific resources
- Target directory`,
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would be downloaded without writing files")
	downloadCmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory (will be prompted if not provided)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	// TODO: Implement the interactive workflow
	// 1. Select cluster(s) from kubeconfig
	// 2. Select namespace(s)
	// 3. Select resource type(s)
	// 4. Optionally select specific resources
	// 5. Select/confirm target directory
	// 6. Fetch and export manifests

	fmt.Println("Download command - to be implemented")
	return nil
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "manifold-k8s",
	Short: "Download Kubernetes manifests from clusters",
	Long: `manifold-k8s is a CLI tool that allows you to interactively select and 
download Kubernetes manifests from one or multiple namespaces.

The tool guides you through:
- Selecting cluster from kubeconfig
- Selecting namespaces
- Selecting resource types (including CRDs, excluding PVs/PVCs)
- Optionally selecting specific manifests
- Choosing target directory`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.toml)")
	rootCmd.PersistentFlags().String("kubeconfig", "", "path to kubeconfig file (default is $HOME/.kube/config)")

	_ = viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("MANIFOLD")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

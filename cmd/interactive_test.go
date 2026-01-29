package cmd

import (
	"testing"

	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestInteractiveCmd(t *testing.T) {
	if interactiveCmd == nil {
		t.Fatal("interactiveCmd is nil")
	}

	if interactiveCmd.Use != "interactive" {
		t.Errorf("interactiveCmd.Use = %s, want interactive", interactiveCmd.Use)
	}

	// Test that command has flags
	flags := []string{"dry-run", "output"}
	for _, flag := range flags {
		f := interactiveCmd.Flag(flag)
		if f == nil {
			t.Errorf("interactiveCmd missing flag: %s", flag)
		}
	}
}

func TestInteractiveCmd_Help(t *testing.T) {
	// Test that --help works
	interactiveCmd.SetArgs([]string{"--help"})
	
	err := interactiveCmd.Execute()
	if err != nil {
		t.Errorf("interactiveCmd.Execute() with --help returned error: %v", err)
	}
}

func TestInteractiveCmd_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{"dry-run flag", "dry-run", "bool"},
		{"output flag", "output", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := interactiveCmd.Flag(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if flag.Value.Type() != tt.wantType {
				t.Errorf("Flag %s type = %s, want %s", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestRunInteractive_LoadKubeConfigError(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()
	
	stubLoadKubeConfig = func(path string) (*api.Config, error) {
		return nil, assert.AnError
	}
	
	// Set up viper
	viper.Set("kubeconfig", "/fake/path")
	
	// Set flags
	interactiveDryRun = false
	interactiveOutputDir = ""
	
	// Run
	err := runInteractive(interactiveCmd, []string{})
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load kubeconfig")
}

func TestRunInteractive_NewClientError(t *testing.T) {
	// Setup
	enableStubs()
	defer disableStubs()
	
	stubNewClient = func(config *api.Config, context string) (*k8s.Client, error) {
		return nil, assert.AnError
	}
	
	// Set up viper
	viper.Set("kubeconfig", "/fake/path")
	
	// Set flags
	interactiveDryRun = false
	interactiveOutputDir = ""
	
	// Run
	err := runInteractive(interactiveCmd, []string{})
	
	// Assert
	// Note: we can't easily test beyond this point without mocking selector prompts
	// The error will occur when trying to prompt for context selection
	assert.Error(t, err)
}

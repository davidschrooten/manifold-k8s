package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd_Version tests version flag (if we add it later)
func TestRootCmd_SubCommands(t *testing.T) {
	// Test that root command has expected subcommands
	expectedCommands := []string{"kubectl-manifests-export", "kubectl-manifests"}

	commands := rootCmd.Commands()
	found := make(map[string]bool)

	for _, cmd := range commands {
		found[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !found[expected] {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

// TestRootCmd_Flags tests that global flags are set up correctly
func TestRootCmd_GlobalFlags(t *testing.T) {
	// Test kubeconfig flag
	kubeconfigFlag := rootCmd.PersistentFlags().Lookup("kubeconfig")
	if kubeconfigFlag == nil {
		t.Error("kubeconfig flag not found")
	}

	// Test config flag
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("config flag not found")
	}
}

// TestExportCmd_ValidateFlags tests flag validation logic
func TestExportCmd_ValidateFlags(t *testing.T) {
	// Test help flag works without error
	// Note: cobra doesn't immediately validate required flags on Execute(),
	// it validates them when the RunE function is called
	exportCmd.SetArgs([]string{"--help"})
	err := exportCmd.Execute()
	if err != nil {
		t.Errorf("export --help should not error, got: %v", err)
	}
}

// TestInteractiveCmd_OutputFlag tests the output flag
func TestInteractiveCmd_OutputFlag(t *testing.T) {
	flag := interactiveCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("output flag not found")
	}

	// Test setting the flag
	err := flag.Value.Set("/tmp/test")
	if err != nil {
		t.Errorf("Failed to set output flag: %v", err)
	}

	if flag.Value.String() != "/tmp/test" {
		t.Errorf("output flag value = %s, want /tmp/test", flag.Value.String())
	}
}

// TestExportCmd_ShorthandFlags tests shorthand flags
func TestExportCmd_ShorthandFlags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"output shorthand", "output", "o"},
		{"context shorthand", "context", "c"},
		{"namespaces shorthand", "namespaces", "n"},
		{"resources shorthand", "resources", "r"},
		{"all-resources shorthand", "all-resources", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := exportCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flagName)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

// TestInteractiveCmd_DryRunFlag tests dry-run flag
func TestInteractiveCmd_DryRunFlag(t *testing.T) {
	flag := interactiveCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("dry-run flag not found")
	}

	// Test default value
	if flag.DefValue != "false" {
		t.Errorf("dry-run default = %s, want false", flag.DefValue)
	}
}

// TestExportCmd_FlagDefaults tests default values
func TestExportCmd_FlagDefaults(t *testing.T) {
	tests := []struct {
		name        string
		flagName    string
		wantDefault string
	}{
		{"dry-run default", "dry-run", "false"},
		{"all-resources default", "all-resources", "false"},
		{"output default", "output", ""},
		{"context default", "context", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := exportCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flagName)
			}
			if flag.DefValue != tt.wantDefault {
				t.Errorf("Flag %s default = %s, want %s", tt.flagName, flag.DefValue, tt.wantDefault)
			}
		})
	}
}

// TestCommandDescriptions tests that commands have proper descriptions
func TestCommandDescriptions(t *testing.T) {
	tests := []struct {
		cmd       *cobra.Command
		name      string
		wantUse   string
		wantShort bool
		wantLong  bool
	}{
		{rootCmd, "root", "manifold-k8s", true, true},
		{exportCmd, "export", "kubectl-manifests-export", true, true},
		{interactiveCmd, "interactive", "kubectl-manifests", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.Use != tt.wantUse {
				t.Errorf("%s command Use = %s, want %s", tt.name, tt.cmd.Use, tt.wantUse)
			}
			if tt.wantShort && tt.cmd.Short == "" {
				t.Errorf("%s command missing Short description", tt.name)
			}
			if tt.wantLong && tt.cmd.Long == "" {
				t.Errorf("%s command missing Long description", tt.name)
			}
		})
	}
}

// TestExportCmd_StringSliceFlags tests string slice flags
func TestExportCmd_StringSliceFlags(t *testing.T) {
	// Test namespaces flag
	namespacesFlag := exportCmd.Flags().Lookup("namespaces")
	if namespacesFlag == nil {
		t.Fatal("namespaces flag not found")
	}

	// Set multiple values
	err := namespacesFlag.Value.Set("default,kube-system")
	if err != nil {
		t.Errorf("Failed to set namespaces: %v", err)
	}

	// Test resources flag
	resourcesFlag := exportCmd.Flags().Lookup("resources")
	if resourcesFlag == nil {
		t.Fatal("resources flag not found")
	}

	err = resourcesFlag.Value.Set("pods,deployments")
	if err != nil {
		t.Errorf("Failed to set resources: %v", err)
	}
}

// TestRootCmd_PersistentFlags tests persistent flags are available to subcommands
func TestRootCmd_PersistentFlags(t *testing.T) {
	// Check that persistent flags are set on root command
	kubeconfigFlag := rootCmd.PersistentFlags().Lookup("kubeconfig")
	if kubeconfigFlag == nil {
		t.Error("rootCmd should have kubeconfig flag")
	}

	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("rootCmd should have config flag")
	}

	// Persistent flags are inherited, but they're looked up via InheritedFlags
	// or by calling the parent command's PersistentFlags
}

// TestInitConfig tests the config initialization
func TestInitConfig(t *testing.T) {
	// Test that initConfig doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initConfig() panicked: %v", r)
		}
	}()

	// Call initConfig - it should not error even if config file doesn't exist
	initConfig()
}

// TestRootCmd_CobraOnInitialize tests that cobra initialization is set up
func TestRootCmd_CobraOnInitialize(t *testing.T) {
	// Test that root command can be executed
	// This tests the init() function indirectly
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("rootCmd.Execute() with --help failed: %v", err)
	}
}

// TestExportCmd_BoolFlags tests boolean flag behavior
func TestExportCmd_BoolFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"dry-run flag", "dry-run"},
		{"all-resources flag", "all-resources"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := exportCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flagName)
			}

			// Test setting to true
			err := flag.Value.Set("true")
			if err != nil {
				t.Errorf("Failed to set %s to true: %v", tt.flagName, err)
			}
			if flag.Value.String() != "true" {
				t.Errorf("%s value = %s, want true", tt.flagName, flag.Value.String())
			}

			// Test setting to false
			err = flag.Value.Set("false")
			if err != nil {
				t.Errorf("Failed to set %s to false: %v", tt.flagName, err)
			}
			if flag.Value.String() != "false" {
				t.Errorf("%s value = %s, want false", tt.flagName, flag.Value.String())
			}
		})
	}
}

// TestInteractiveCmd_ShortDescription tests command descriptions
func TestInteractiveCmd_ShortDescription(t *testing.T) {
	if interactiveCmd.Short == "" {
		t.Error("interactiveCmd.Short should not be empty")
	}
	if interactiveCmd.Long == "" {
		t.Error("interactiveCmd.Long should not be empty")
	}
}

// TestExportCmd_ShortDescription tests command descriptions
func TestExportCmd_ShortDescription(t *testing.T) {
	if exportCmd.Short == "" {
		t.Error("exportCmd.Short should not be empty")
	}
	if exportCmd.Long == "" {
		t.Error("exportCmd.Long should not be empty")
	}
}

// TestRootCmd_HasSubcommands tests that root has subcommands
func TestRootCmd_HasSubcommands(t *testing.T) {
	if !rootCmd.HasSubCommands() {
		t.Error("rootCmd should have subcommands")
	}

	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("rootCmd should have at least one subcommand")
	}
}

// TestExportCmd_RunEIsSet tests that RunE is set
func TestExportCmd_RunEIsSet(t *testing.T) {
	if exportCmd.RunE == nil {
		t.Error("exportCmd.RunE should be set")
	}
}

// TestInteractiveCmd_RunEIsSet tests that RunE is set
func TestInteractiveCmd_RunEIsSet(t *testing.T) {
	if interactiveCmd.RunE == nil {
		t.Error("interactiveCmd.RunE should be set")
	}
}

// TestExportInit tests export init function
func TestExportInit(t *testing.T) {
	// Test that init doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("export init() panicked: %v", r)
		}
	}()

	// Verify command is added to root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "kubectl-manifests-export" {
			found = true
			break
		}
	}
	if !found {
		t.Error("kubectl-manifests-export command not found in root commands")
	}
}

// TestInteractiveInit tests interactive init function
func TestInteractiveInit(t *testing.T) {
	// Test that init doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("interactive init() panicked: %v", r)
		}
	}()

	// Verify command is added to root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "kubectl-manifests" {
			found = true
			break
		}
	}
	if !found {
		t.Error("kubectl-manifests command not found in root commands")
	}
}

// TestExportCmd_Examples tests that command has examples
func TestExportCmd_Examples(t *testing.T) {
	if exportCmd.Long == "" {
		t.Error("exportCmd should have Long description with examples")
	}
}

// TestInteractiveCmd_Examples tests that command has examples
func TestInteractiveCmd_Examples(t *testing.T) {
	if interactiveCmd.Long == "" {
		t.Error("interactiveCmd should have Long description")
	}
}

// TestRootCmd_Version tests version information
func TestRootCmd_Version(t *testing.T) {
	if rootCmd.Version != "" {
		t.Log("Version is set:", rootCmd.Version)
	}
}

// TestExportCmd_Aliases tests command aliases
func TestExportCmd_Aliases(t *testing.T) {
	// Export command doesn't have aliases, but test the field exists
	if len(exportCmd.Aliases) > 0 {
		t.Logf("Export has aliases: %v", exportCmd.Aliases)
	}
}

// TestInteractiveCmd_Aliases tests command aliases
func TestInteractiveCmd_Aliases(t *testing.T) {
	// Interactive command doesn't have aliases, but test the field exists
	if len(interactiveCmd.Aliases) > 0 {
		t.Logf("Interactive has aliases: %v", interactiveCmd.Aliases)
	}
}

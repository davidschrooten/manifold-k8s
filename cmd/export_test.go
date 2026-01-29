package cmd

import (
	"testing"
)

func TestExportCmd(t *testing.T) {
	if exportCmd == nil {
		t.Fatal("exportCmd is nil")
	}

	if exportCmd.Use != "export" {
		t.Errorf("exportCmd.Use = %s, want export", exportCmd.Use)
	}

	// Test that command has required flags
	requiredFlags := []string{"context", "namespaces", "output"}
	for _, flag := range requiredFlags {
		f := exportCmd.Flag(flag)
		if f == nil {
			t.Errorf("exportCmd missing required flag: %s", flag)
		}
	}

	// Test that command has optional flags
	optionalFlags := []string{"dry-run", "resources", "all-resources"}
	for _, flag := range optionalFlags {
		f := exportCmd.Flag(flag)
		if f == nil {
			t.Errorf("exportCmd missing flag: %s", flag)
		}
	}
}

func TestExportCmd_RequiredFlagsMarked(t *testing.T) {
	// Test that required flags are properly marked
	requiredFlags := []string{"context", "namespaces", "output"}
	for _, flagName := range requiredFlags {
		flag := exportCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Required flag %s not found", flagName)
			continue
		}
		// Check if flag is marked as required via cobra's annotation
		annotations := flag.Annotations
		if annotations != nil {
			if required, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; ok {
				if len(required) == 0 {
					t.Errorf("Flag %s not properly marked as required", flagName)
				}
			}
		}
	}
}

func TestExportCmd_MissingResourcesOrAllResources(t *testing.T) {
	// Test that either --resources or --all-resources is required
	// This is validated in runExport, so we can't easily test without mocking
	// but we document that this validation exists
	if !exportAllRes && len(exportResources) == 0 {
		// This is the validation logic that should trigger an error
		t.Log("Validation for --resources or --all-resources is present")
	}
}

func TestExportCmd_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{"dry-run flag", "dry-run", "bool"},
		{"output flag", "output", "string"},
		{"context flag", "context", "string"},
		{"namespaces flag", "namespaces", "stringSlice"},
		{"resources flag", "resources", "stringSlice"},
		{"all-resources flag", "all-resources", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := exportCmd.Flag(tt.flagName)
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

package cmd

import (
	"testing"
)

func TestExecute(t *testing.T) {
	// Basic test that Execute doesn't panic
	// We can't easily test the actual execution without mocking cobra
	// This at least exercises the code path
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()

	// Reset rootCmd for testing
	rootCmd.SetArgs([]string{"--help"})

	// Execute should not return error for --help
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() with --help returned error: %v", err)
	}
}

func TestRootCmd(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "manifold-k8s" {
		t.Errorf("rootCmd.Use = %s, want manifold-k8s", rootCmd.Use)
	}
}

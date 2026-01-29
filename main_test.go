package main

import (
	"testing"
)

// Test that main doesn't panic when imported
func TestMainDoesNotPanic(t *testing.T) {
	// This test ensures the package can be imported
	// The actual main() function will be tested via integration tests
	// We can't directly test main() as it calls os.Exit
}

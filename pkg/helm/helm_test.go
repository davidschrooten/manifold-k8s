package helm

import (
	"os"
	"os/exec"
	"testing"
)

func TestIsHelmInstalled(t *testing.T) {
	// This test depends on whether helm is actually installed
	// We'll test both scenarios by manipulating PATH

	originalPath := os.Getenv("PATH")
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	t.Run("helm not in PATH", func(t *testing.T) {
		_ = os.Setenv("PATH", "/nonexistent")
		if IsHelmInstalled() {
			t.Error("IsHelmInstalled() should return false when helm is not in PATH")
		}
	})

	t.Run("helm in PATH", func(t *testing.T) {
		_ = os.Setenv("PATH", originalPath)
		// Look for helm in the original PATH
		_, err := exec.LookPath("helm")
		if err == nil {
			// helm is installed
			if !IsHelmInstalled() {
				t.Error("IsHelmInstalled() should return true when helm is in PATH")
			}
		} else {
			// helm is not installed, skip this test
			t.Skip("helm is not installed, skipping positive test")
		}
	})
}

func TestListReleases_NoHelm(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to simulate missing helm
	_ = os.Setenv("PATH", "/nonexistent")

	_, err := ListReleases("default")
	if err == nil {
		t.Error("ListReleases() should return error when helm is not available")
	}
}

func TestListReleases_InvalidNamespace(t *testing.T) {
	// Check if helm is installed
	if !IsHelmInstalled() {
		t.Skip("helm is not installed, skipping test")
	}

	// Try to list releases in a namespace that likely doesn't exist
	_, err := ListReleases("nonexistent-namespace-12345")
	// This should not error - it should just return empty list
	// Helm list returns success with empty results for non-existent namespaces
	if err != nil {
		t.Logf("ListReleases() returned error: %v (this is acceptable)", err)
	}
}

func TestGetValues_NoHelm(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to simulate missing helm
	_ = os.Setenv("PATH", "/nonexistent")

	_, err := GetValues("test-release", "default")
	if err == nil {
		t.Error("GetValues() should return error when helm is not available")
	}
}

func TestGetValues_NonExistentRelease(t *testing.T) {
	// Check if helm is installed
	if !IsHelmInstalled() {
		t.Skip("helm is not installed, skipping test")
	}

	// Try to get values for a release that doesn't exist
	_, err := GetValues("nonexistent-release-12345", "default")
	if err == nil {
		t.Error("GetValues() should return error for non-existent release")
	}

	if err != nil && err.Error() == "" {
		t.Error("GetValues() error should have a message")
	}
}

func TestRelease_Structure(t *testing.T) {
	// Test that Release struct can be created and has expected fields
	release := Release{
		Name:      "test-release",
		Namespace: "default",
		Chart:     "nginx-1.0.0",
		Status:    "deployed",
		Revision:  "1",
	}

	if release.Name != "test-release" {
		t.Errorf("Release.Name = %s, want test-release", release.Name)
	}
	if release.Namespace != "default" {
		t.Errorf("Release.Namespace = %s, want default", release.Namespace)
	}
	if release.Chart != "nginx-1.0.0" {
		t.Errorf("Release.Chart = %s, want nginx-1.0.0", release.Chart)
	}
	if release.Status != "deployed" {
		t.Errorf("Release.Status = %s, want deployed", release.Status)
	}
	if release.Revision != "1" {
		t.Errorf("Release.Revision = %s, want 1", release.Revision)
	}
}

func TestListReleasesWithContext_NoHelm(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to simulate missing helm
	_ = os.Setenv("PATH", "/nonexistent")

	_, err := ListReleasesWithContext("default", "test-context")
	if err == nil {
		t.Error("ListReleasesWithContext() should return error when helm is not available")
	}
}

func TestGetValuesWithContext_NoHelm(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to simulate missing helm
	_ = os.Setenv("PATH", "/nonexistent")

	_, err := GetValuesWithContext("test-release", "default", "test-context")
	if err == nil {
		t.Error("GetValuesWithContext() should return error when helm is not available")
	}
}

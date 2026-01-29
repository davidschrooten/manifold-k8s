package helm

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Release represents a Helm release
type Release struct {
	Name      string
	Namespace string
	Chart     string
	Status    string
	Revision  string
}

// IsHelmInstalled checks if helm CLI is available
func IsHelmInstalled() bool {
	cmd := exec.Command("helm", "version", "--short")
	return cmd.Run() == nil
}

// ListReleases lists all Helm releases in a namespace using the specified context
func ListReleases(namespace string) ([]Release, error) {
	return ListReleasesWithContext(namespace, "")
}

// ListReleasesWithContext lists all Helm releases in a namespace using the specified kube context
func ListReleasesWithContext(namespace, kubeContext string) ([]Release, error) {
	args := []string{"list", "-n", namespace}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	args = append(args, "--output", "json")

	cmd := exec.Command("helm", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list helm releases: %s", stderr.String())
	}

	// Parse JSON output (simplified - just parse the text output instead)
	// Using --output table for easier parsing
	args = []string{"list", "-n", namespace}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	cmd = exec.Command("helm", args...)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list helm releases: %s", stderr.String())
	}

	output := stdout.String()
	lines := strings.Split(output, "\n")

	var releases []Release
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			releases = append(releases, Release{
				Name:      fields[0],
				Namespace: fields[1],
				Revision:  fields[2],
				Status:    fields[7],
				Chart:     fields[8],
			})
		}
	}

	return releases, nil
}

// GetValues retrieves the values for a specific Helm release
func GetValues(releaseName, namespace string) (string, error) {
	return GetValuesWithContext(releaseName, namespace, "")
}

// GetValuesWithContext retrieves the values for a specific Helm release using the specified kube context
func GetValuesWithContext(releaseName, namespace, kubeContext string) (string, error) {
	args := []string{"get", "values", releaseName, "-n", namespace, "--all"}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	cmd := exec.Command("helm", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get helm values for %s: %s", releaseName, stderr.String())
	}

	return stdout.String(), nil
}

package selector

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
)

// FormatContextOptions formats context names, marking the current one
func FormatContextOptions(contexts []string, currentContext string) []string {
	options := make([]string, len(contexts))
	for i, ctx := range contexts {
		if ctx == currentContext {
			options[i] = fmt.Sprintf("%s (current)", ctx)
		} else {
			options[i] = ctx
		}
	}
	return options
}

// FormatResourceOptions formats resource info into display strings
func FormatResourceOptions(resources []k8s.ResourceInfo) []string {
	options := make([]string, len(resources))
	for i, res := range resources {
		options[i] = res.String()
	}
	return options
}

// ParseSelectedContext extracts the context name from a formatted option
func ParseSelectedContext(selected string) string {
	// Remove " (current)" suffix if present
	return strings.TrimSuffix(selected, " (current)")
}

// ParseSelectedResource finds the resource info from a formatted option
func ParseSelectedResource(selected string, resources []k8s.ResourceInfo) *k8s.ResourceInfo {
	for _, res := range resources {
		if res.String() == selected {
			return &res
		}
	}
	return nil
}

// ValidateDirectory validates that a directory path is not empty
func ValidateDirectory(input interface{}) error {
	str, ok := input.(string)
	if !ok {
		return errors.New("invalid input type")
	}
	if strings.TrimSpace(str) == "" {
		return errors.New("directory path cannot be empty")
	}
	return nil
}

// PromptContextSelection prompts the user to select one or more contexts
func PromptContextSelection(contexts []string, currentContext string) ([]string, error) {
	return promptContextSelectionWithAsker(askOne, contexts, currentContext)
}

func promptContextSelectionWithAsker(asker AskOneFunc, contexts []string, currentContext string) ([]string, error) {
	if len(contexts) == 0 {
		return nil, errors.New("no contexts available")
	}

	options := FormatContextOptions(contexts, currentContext)
	
	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select cluster context(s):",
		Options: options,
		Default: []string{fmt.Sprintf("%s (current)", currentContext)},
	}
	
	if err := asker(prompt, &selected); err != nil {
		return nil, fmt.Errorf("failed to select contexts: %w", err)
	}

	// Parse selected contexts
	result := make([]string, len(selected))
	for i, sel := range selected {
		result[i] = ParseSelectedContext(sel)
	}

	return result, nil
}

// PromptNamespaceSelection prompts the user to select one or more namespaces
func PromptNamespaceSelection(namespaces []string) ([]string, error) {
	return promptNamespaceSelectionWithAsker(askOne, namespaces)
}

func promptNamespaceSelectionWithAsker(asker AskOneFunc, namespaces []string) ([]string, error) {
	if len(namespaces) == 0 {
		return nil, errors.New("no namespaces available")
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select namespace(s):",
		Options: namespaces,
	}
	
	if err := asker(prompt, &selected); err != nil {
		return nil, fmt.Errorf("failed to select namespaces: %w", err)
	}

	if len(selected) == 0 {
		return nil, errors.New("no namespaces selected")
	}

	return selected, nil
}

// PromptResourceSelection prompts the user to select one or more resource types
func PromptResourceSelection(resources []k8s.ResourceInfo) ([]k8s.ResourceInfo, error) {
	return promptResourceSelectionWithAsker(askOne, resources)
}

func promptResourceSelectionWithAsker(asker AskOneFunc, resources []k8s.ResourceInfo) ([]k8s.ResourceInfo, error) {
	if len(resources) == 0 {
		return nil, errors.New("no resources available")
	}

	options := FormatResourceOptions(resources)
	
	var selected []string
	prompt := &survey.MultiSelect{
		Message:  "Select resource type(s):",
		Options:  options,
		PageSize: 15,
	}
	
	if err := asker(prompt, &selected); err != nil {
		return nil, fmt.Errorf("failed to select resources: %w", err)
	}

	if len(selected) == 0 {
		return nil, errors.New("no resources selected")
	}

	// Parse selected resources
	result := make([]k8s.ResourceInfo, 0, len(selected))
	for _, sel := range selected {
		if res := ParseSelectedResource(sel, resources); res != nil {
			result = append(result, *res)
		}
	}

	return result, nil
}

// PromptDirectorySelection prompts the user for a target directory
func PromptDirectorySelection(defaultDir string) (string, error) {
	return promptDirectorySelectionWithAsker(askOne, defaultDir)
}

func promptDirectorySelectionWithAsker(asker AskOneFunc, defaultDir string) (string, error) {
	var directory string
	prompt := &survey.Input{
		Message: "Target directory for manifests:",
		Default: defaultDir,
	}
	
	if err := asker(prompt, &directory, survey.WithValidator(ValidateDirectory)); err != nil {
		return "", fmt.Errorf("failed to get directory: %w", err)
	}

	return directory, nil
}

// PromptConfirmation prompts the user for a yes/no confirmation
func PromptConfirmation(message string) (bool, error) {
	return promptConfirmationWithAsker(askOne, message)
}

func promptConfirmationWithAsker(asker AskOneFunc, message string) (bool, error) {
	var confirmed bool
	prompt := &survey.Confirm{
		Message: message,
		Default: false,
	}
	
	if err := asker(prompt, &confirmed); err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}

	return confirmed, nil
}

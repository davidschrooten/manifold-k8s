package selector

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidschrooten/manifold-k8s/pkg/k8s"
)

// Prompter is an interface for prompting the user
// This allows for mocking in tests
type Prompter interface {
	PromptContextSelection(contexts []string, currentContext string) ([]string, error)
	PromptNamespaceSelection(namespaces []string) ([]string, error)
	PromptResourceSelection(resources []k8s.ResourceInfo) ([]k8s.ResourceInfo, error)
	PromptDirectorySelection(defaultDir string) (string, error)
	PromptConfirmation(message string) (bool, error)
}

// DefaultPrompter uses the survey library for actual prompts
type DefaultPrompter struct{}

// NewDefaultPrompter creates a new default prompter
func NewDefaultPrompter() *DefaultPrompter {
	return &DefaultPrompter{}
}

// PromptContextSelection implements Prompter
func (p *DefaultPrompter) PromptContextSelection(contexts []string, currentContext string) ([]string, error) {
	return PromptContextSelection(contexts, currentContext)
}

// PromptNamespaceSelection implements Prompter
func (p *DefaultPrompter) PromptNamespaceSelection(namespaces []string) ([]string, error) {
	return PromptNamespaceSelection(namespaces)
}

// PromptResourceSelection implements Prompter
func (p *DefaultPrompter) PromptResourceSelection(resources []k8s.ResourceInfo) ([]k8s.ResourceInfo, error) {
	return PromptResourceSelection(resources)
}

// PromptDirectorySelection implements Prompter
func (p *DefaultPrompter) PromptDirectorySelection(defaultDir string) (string, error) {
	return PromptDirectorySelection(defaultDir)
}

// PromptConfirmation implements Prompter
func (p *DefaultPrompter) PromptConfirmation(message string) (bool, error) {
	return PromptConfirmation(message)
}

// AskOneFunc is a function type that matches survey.AskOne signature
type AskOneFunc func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error

// SetAskOne allows overriding the survey.AskOne function for testing
var askOne AskOneFunc = survey.AskOne

// PromptWithAsker allows testing by injecting a custom asker
func PromptWithAsker(asker AskOneFunc, prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	return asker(prompt, response, opts...)
}

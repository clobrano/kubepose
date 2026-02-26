package tui

import (
	"github.com/clobrano/kubepose/internal/kubectl"
)

// ContextLoadedMsg is sent when the kubectl context info is loaded
type ContextLoadedMsg struct {
	Context   string
	Namespace string
}

// ResourcesLoadedMsg is sent when resources are loaded from kubectl
type ResourcesLoadedMsg struct {
	Data *kubectl.TableData
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err       error
	Retryable bool
}

// RefreshMsg requests a refresh of the current view
type RefreshMsg struct{}

// TabChangedMsg is sent when the active tab changes
type TabChangedMsg struct {
	Index int
}

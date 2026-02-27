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

// DetailLoadedMsg is sent when resource detail is loaded
type DetailLoadedMsg struct {
	ResourceName string
	Content      string
	Format       int // 0=Table, 1=YAML, 2=JSON
}

// DescribeLoadedMsg is sent when describe output is loaded
type DescribeLoadedMsg struct {
	ResourceNames []string
	Content       string
}

// ContainersLoadedMsg is sent when a pod's containers are retrieved
type ContainersLoadedMsg struct {
	PodName    string
	Namespace  string
	Containers []string
	Follow     bool // true if this is for follow mode logs
}

// LogsLoadedMsg is sent when log output is loaded
type LogsLoadedMsg struct {
	PodName   string
	Container string
	Content   string
}

// ExecRequestMsg is sent to request exec into a container
type ExecRequestMsg struct {
	PodName   string
	Namespace string
	Container string
}

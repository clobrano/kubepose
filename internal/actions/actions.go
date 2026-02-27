package actions

import (
	"github.com/clobrano/kubepose/internal/kubectl"
)

// ActionContext contains the context for executing an action
type ActionContext struct {
	Kubectl      *kubectl.Kubectl
	ResourceType string
	Names        []string
	Namespace    string
	Container    string
}

// NewContext creates a new action context
func NewContext(k *kubectl.Kubectl, resourceType string, names []string, namespace string) *ActionContext {
	return &ActionContext{
		Kubectl:      k,
		ResourceType: resourceType,
		Names:        names,
		Namespace:    namespace,
	}
}

// WithContainer returns a copy of the context with a container specified
func (c *ActionContext) WithContainer(container string) *ActionContext {
	return &ActionContext{
		Kubectl:      c.Kubectl,
		ResourceType: c.ResourceType,
		Names:        c.Names,
		Namespace:    c.Namespace,
		Container:    container,
	}
}

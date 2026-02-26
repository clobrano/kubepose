package header

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	h := New("my-context", "my-namespace", 80)

	if h.Context() != "my-context" {
		t.Errorf("Context() = %q, want %q", h.Context(), "my-context")
	}

	if h.Namespace() != "my-namespace" {
		t.Errorf("Namespace() = %q, want %q", h.Namespace(), "my-namespace")
	}
}

func TestSetContext(t *testing.T) {
	h := New("old-context", "ns", 80)
	h.SetContext("new-context")

	if h.Context() != "new-context" {
		t.Errorf("Context() = %q, want %q", h.Context(), "new-context")
	}
}

func TestSetNamespace(t *testing.T) {
	h := New("ctx", "old-ns", 80)
	h.SetNamespace("new-ns")

	if h.Namespace() != "new-ns" {
		t.Errorf("Namespace() = %q, want %q", h.Namespace(), "new-ns")
	}
}

func TestViewContainsContext(t *testing.T) {
	h := New("test-cluster", "default", 100)
	view := h.View()

	if !strings.Contains(view, "Context:") {
		t.Error("View should contain 'Context:'")
	}

	if !strings.Contains(view, "test-cluster") {
		t.Error("View should contain context name")
	}
}

func TestViewContainsNamespace(t *testing.T) {
	h := New("ctx", "kube-system", 100)
	view := h.View()

	if !strings.Contains(view, "Namespace:") {
		t.Error("View should contain 'Namespace:'")
	}

	if !strings.Contains(view, "kube-system") {
		t.Error("View should contain namespace name")
	}
}

func TestViewContainsHelp(t *testing.T) {
	h := New("ctx", "ns", 100)
	view := h.View()

	if !strings.Contains(view, "[?] Help") {
		t.Error("View should contain '[?] Help'")
	}
}

func TestViewEmptyWidth(t *testing.T) {
	h := New("ctx", "ns", 0)
	view := h.View()

	if view != "" {
		t.Errorf("View with width 0 should be empty, got %q", view)
	}
}

func TestViewEmptyContext(t *testing.T) {
	h := New("", "ns", 80)
	view := h.View()

	if !strings.Contains(view, "(none)") {
		t.Error("View should show '(none)' for empty context")
	}
}

func TestViewEmptyNamespace(t *testing.T) {
	h := New("ctx", "", 80)
	view := h.View()

	if !strings.Contains(view, "default") {
		t.Error("View should show 'default' for empty namespace")
	}
}

func TestString(t *testing.T) {
	h := New("ctx", "ns", 80)
	s := h.String()

	if !strings.Contains(s, "ctx") || !strings.Contains(s, "ns") {
		t.Errorf("String() = %q, should contain context and namespace", s)
	}
}

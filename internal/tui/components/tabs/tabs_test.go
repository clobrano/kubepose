package tabs

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tabs := New([]string{"Pods", "Deployments", "Services"}, 0)

	if tabs.Active() != 0 {
		t.Errorf("Active() = %d, want 0", tabs.Active())
	}

	if tabs.Count() != 3 {
		t.Errorf("Count() = %d, want 3", tabs.Count())
	}
}

func TestNewWithInvalidIndex(t *testing.T) {
	// Negative index should be clamped to 0
	tabs := New([]string{"A", "B"}, -1)
	if tabs.Active() != 0 {
		t.Errorf("Active() with negative index = %d, want 0", tabs.Active())
	}

	// Index beyond range should be clamped to last
	tabs = New([]string{"A", "B"}, 10)
	if tabs.Active() != 1 {
		t.Errorf("Active() with out-of-range index = %d, want 1", tabs.Active())
	}
}

func TestSetActive(t *testing.T) {
	tabs := New([]string{"A", "B", "C"}, 0)

	tabs.SetActive(1)
	if tabs.Active() != 1 {
		t.Errorf("Active() after SetActive(1) = %d, want 1", tabs.Active())
	}

	// Invalid indices should be ignored
	tabs.SetActive(-1)
	if tabs.Active() != 1 {
		t.Errorf("Active() after SetActive(-1) = %d, want 1", tabs.Active())
	}

	tabs.SetActive(100)
	if tabs.Active() != 1 {
		t.Errorf("Active() after SetActive(100) = %d, want 1", tabs.Active())
	}
}

func TestActiveName(t *testing.T) {
	tabs := New([]string{"Pods", "Deployments"}, 0)

	if tabs.ActiveName() != "Pods" {
		t.Errorf("ActiveName() = %q, want %q", tabs.ActiveName(), "Pods")
	}

	tabs.SetActive(1)
	if tabs.ActiveName() != "Deployments" {
		t.Errorf("ActiveName() = %q, want %q", tabs.ActiveName(), "Deployments")
	}
}

func TestNext(t *testing.T) {
	tabs := New([]string{"A", "B", "C"}, 0)

	tabs.Next()
	if tabs.Active() != 1 {
		t.Errorf("Active() after Next() = %d, want 1", tabs.Active())
	}

	tabs.Next()
	if tabs.Active() != 2 {
		t.Errorf("Active() after second Next() = %d, want 2", tabs.Active())
	}

	// Should wrap around
	tabs.Next()
	if tabs.Active() != 0 {
		t.Errorf("Active() after wrap-around Next() = %d, want 0", tabs.Active())
	}
}

func TestPrevious(t *testing.T) {
	tabs := New([]string{"A", "B", "C"}, 2)

	tabs.Previous()
	if tabs.Active() != 1 {
		t.Errorf("Active() after Previous() = %d, want 1", tabs.Active())
	}

	tabs.Previous()
	if tabs.Active() != 0 {
		t.Errorf("Active() after second Previous() = %d, want 0", tabs.Active())
	}

	// Should wrap around
	tabs.Previous()
	if tabs.Active() != 2 {
		t.Errorf("Active() after wrap-around Previous() = %d, want 2", tabs.Active())
	}
}

func TestNextPreviousEmpty(t *testing.T) {
	tabs := New([]string{}, 0)

	// Should not panic
	tabs.Next()
	tabs.Previous()
}

func TestViewContainsTabs(t *testing.T) {
	tabs := New([]string{"Pods", "Services"}, 0)
	view := tabs.View()

	if !strings.Contains(view, "Pods") {
		t.Error("View should contain 'Pods'")
	}

	if !strings.Contains(view, "Services") {
		t.Error("View should contain 'Services'")
	}
}

func TestViewContainsNumbers(t *testing.T) {
	tabs := New([]string{"A", "B", "C"}, 0)
	view := tabs.View()

	if !strings.Contains(view, "1") {
		t.Error("View should contain tab number '1'")
	}

	if !strings.Contains(view, "2") {
		t.Error("View should contain tab number '2'")
	}

	if !strings.Contains(view, "3") {
		t.Error("View should contain tab number '3'")
	}
}

func TestViewEmpty(t *testing.T) {
	tabs := New([]string{}, 0)
	view := tabs.View()

	if view != "" {
		t.Errorf("View with no tabs should be empty, got %q", view)
	}
}

func TestString(t *testing.T) {
	tabs := New([]string{"A", "B"}, 1)
	s := tabs.String()

	if !strings.Contains(s, "active=1") {
		t.Errorf("String() = %q, should contain 'active=1'", s)
	}
}

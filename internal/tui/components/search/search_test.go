package search

import (
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		text      string
		wantMatch bool
	}{
		{"empty pattern matches anything", "", "hello", true},
		{"exact match", "hello", "hello", true},
		{"case insensitive", "HELLO", "hello", true},
		{"case insensitive reverse", "hello", "HELLO", true},
		{"prefix match", "hel", "hello", true},
		{"suffix match", "llo", "hello", true},
		{"middle match", "ell", "hello", true},
		{"fuzzy match", "hlo", "hello", true},
		{"fuzzy match scattered", "nxabc", "nginx-abc123", true},
		{"no match", "xyz", "hello", false},
		{"pattern longer than text", "hello world", "hello", false},
		{"partial pattern match fails", "hellox", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, _ := FuzzyMatch(tt.pattern, tt.text)
			if match != tt.wantMatch {
				t.Errorf("FuzzyMatch(%q, %q) = %v, want %v", tt.pattern, tt.text, match, tt.wantMatch)
			}
		})
	}
}

func TestFuzzyMatchScore(t *testing.T) {
	// Exact match should score highest
	_, exactScore := FuzzyMatch("nginx", "nginx")
	_, prefixScore := FuzzyMatch("nginx", "nginx-deployment")
	_, fuzzyScore := FuzzyMatch("ngx", "nginx")

	if exactScore <= prefixScore {
		t.Errorf("Exact match score (%d) should be > prefix match score (%d)", exactScore, prefixScore)
	}

	if prefixScore <= fuzzyScore {
		t.Errorf("Prefix match score (%d) should be > fuzzy match score (%d)", prefixScore, fuzzyScore)
	}
}

func TestFuzzyMatchRow(t *testing.T) {
	row := []string{"nginx-7c6b7d8b9-abc12", "Running", "default"}

	// Should match name
	match, _ := FuzzyMatchRow("nginx", row)
	if !match {
		t.Error("Should match 'nginx' in row")
	}

	// Should match status
	match, _ = FuzzyMatchRow("running", row)
	if !match {
		t.Error("Should match 'running' in row")
	}

	// Should match namespace
	match, _ = FuzzyMatchRow("default", row)
	if !match {
		t.Error("Should match 'default' in row")
	}

	// Should not match
	match, _ = FuzzyMatchRow("redis", row)
	if match {
		t.Error("Should not match 'redis' in row")
	}
}

func TestFilterRows(t *testing.T) {
	rows := [][]string{
		{"nginx-abc", "Running"},
		{"redis-xyz", "Pending"},
		{"nginx-def", "Running"},
		{"postgres-123", "Running"},
	}

	// Filter for nginx
	filtered := FilterRows("nginx", rows)
	if len(filtered) != 2 {
		t.Errorf("FilterRows('nginx') returned %d rows, want 2", len(filtered))
	}

	// Filter for running
	filtered = FilterRows("running", rows)
	if len(filtered) != 3 {
		t.Errorf("FilterRows('running') returned %d rows, want 3", len(filtered))
	}

	// Empty filter returns all
	filtered = FilterRows("", rows)
	if len(filtered) != 4 {
		t.Errorf("FilterRows('') returned %d rows, want 4", len(filtered))
	}

	// No matches returns empty
	filtered = FilterRows("nonexistent", rows)
	if len(filtered) != 0 {
		t.Errorf("FilterRows('nonexistent') returned %d rows, want 0", len(filtered))
	}
}

func TestSearchActivateDeactivate(t *testing.T) {
	s := New()

	if s.IsActive() {
		t.Error("Search should not be active initially")
	}

	s.Activate()
	if !s.IsActive() {
		t.Error("Search should be active after Activate()")
	}

	s.Deactivate()
	if s.IsActive() {
		t.Error("Search should not be active after Deactivate()")
	}
}

func TestSearchIsFiltered(t *testing.T) {
	s := New()

	if s.IsFiltered() {
		t.Error("Search should not be filtered initially")
	}

	s.Activate()
	s.Filter("nginx")

	if !s.IsFiltered() {
		t.Error("Search should be filtered after Filter() with non-empty query")
	}

	// Deactivate clears the filter
	s.Deactivate()
	if s.IsFiltered() {
		t.Error("Search should not be filtered after Deactivate()")
	}
}

func TestSearchViewFilteredButInactive(t *testing.T) {
	s := New()
	rows := [][]string{
		{"nginx-abc", "Running"},
		{"redis-xyz", "Pending"},
	}

	s.SetItems(rows)
	s.Activate()
	s.Filter("nginx")

	// Simulate Enter: deactivate typing but keep filter
	s.active = false

	view := s.View()
	if view == "" {
		t.Error("View() should not be empty when filter is applied (even if not in active typing mode)")
	}
}

func TestSearchSetItems(t *testing.T) {
	s := New()
	rows := [][]string{
		{"nginx", "Running"},
		{"redis", "Pending"},
	}

	s.SetItems(rows)
	filtered := s.FilteredItems()

	if len(filtered) != 2 {
		t.Errorf("FilteredItems() returned %d rows, want 2", len(filtered))
	}
}

func TestSearchFilter(t *testing.T) {
	s := New()
	rows := [][]string{
		{"nginx-abc", "Running"},
		{"redis-xyz", "Pending"},
		{"nginx-def", "Running"},
	}

	s.SetItems(rows)
	s.Activate()
	s.Filter("nginx")

	filtered := s.FilteredItems()
	if len(filtered) != 2 {
		t.Errorf("FilteredItems() after Filter('nginx') returned %d rows, want 2", len(filtered))
	}
}

func TestSearchQuery(t *testing.T) {
	s := New()

	if s.Query() != "" {
		t.Errorf("Query() on new search = %q, want empty", s.Query())
	}
}

func TestSearchViewInactive(t *testing.T) {
	s := New()
	view := s.View()

	if view != "" {
		t.Errorf("View() when inactive should be empty, got %q", view)
	}
}

func TestSearchViewActive(t *testing.T) {
	s := New()
	s.Activate()
	view := s.View()

	if view == "" {
		t.Error("View() when active should not be empty")
	}
}

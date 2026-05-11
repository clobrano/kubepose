package kubectl

import (
	"testing"
	"time"
)

func TestParseAge(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"5s", 5 * time.Second},
		{"2m", 2 * time.Minute},
		{"2m30s", 2*time.Minute + 30*time.Second},
		{"1h", time.Hour},
		{"1h30m", 90 * time.Minute},
		{"2d", 48 * time.Hour},
		{"2d3h", 48*time.Hour + 3*time.Hour},
		{"10d2h30m", 10*24*time.Hour + 2*time.Hour + 30*time.Minute},
		{"", 0},
		{"<unknown>", 0},
		{"-", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseAge(tt.input)
			if got != tt.want {
				t.Errorf("parseAge(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTableDataSort_ByName(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "STATUS", "AGE"},
		Rows: [][]string{
			{"zebra", "Running", "1h"},
			{"apple", "Pending", "2h"},
			{"mango", "Running", "30m"},
		},
	}

	data.Sort("NAME", false)
	wantOrder := []string{"apple", "mango", "zebra"}
	for i, want := range wantOrder {
		if data.Rows[i][0] != want {
			t.Errorf("row %d name = %q, want %q", i, data.Rows[i][0], want)
		}
	}
}

func TestTableDataSort_ByNameReverse(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "STATUS", "AGE"},
		Rows: [][]string{
			{"zebra", "Running", "1h"},
			{"apple", "Pending", "2h"},
			{"mango", "Running", "30m"},
		},
	}

	data.Sort("NAME", true)
	wantOrder := []string{"zebra", "mango", "apple"}
	for i, want := range wantOrder {
		if data.Rows[i][0] != want {
			t.Errorf("row %d name = %q, want %q", i, data.Rows[i][0], want)
		}
	}
}

func TestTableDataSort_ByAge(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "AGE"},
		Rows: [][]string{
			{"old", "10d"},
			{"medium", "2h"},
			{"new", "5m"},
		},
	}

	// ascending = newest (shortest age) first
	data.Sort("AGE", false)
	wantOrder := []string{"new", "medium", "old"}
	for i, want := range wantOrder {
		if data.Rows[i][0] != want {
			t.Errorf("row %d name = %q, want %q", i, data.Rows[i][0], want)
		}
	}
}

func TestTableDataSort_ByCreationTime(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "AGE"},
		Rows: [][]string{
			{"old", "10d"},
			{"medium", "2h"},
			{"new", "5m"},
		},
	}

	// creation_time ascending = newest first (smallest age value)
	data.Sort("creation_time", false)
	wantOrder := []string{"new", "medium", "old"}
	for i, want := range wantOrder {
		if data.Rows[i][0] != want {
			t.Errorf("row %d name = %q, want %q", i, data.Rows[i][0], want)
		}
	}
}

func TestTableDataSort_ByCreationTimeReverse(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "AGE"},
		Rows: [][]string{
			{"old", "10d"},
			{"medium", "2h"},
			{"new", "5m"},
		},
	}

	// creation_time reverse = oldest first (largest age value)
	data.Sort("creation_time", true)
	wantOrder := []string{"old", "medium", "new"}
	for i, want := range wantOrder {
		if data.Rows[i][0] != want {
			t.Errorf("row %d name = %q, want %q", i, data.Rows[i][0], want)
		}
	}
}

func TestTableDataSort_UnknownColumn(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"zebra", "Running"},
			{"apple", "Pending"},
		},
	}

	// Should not panic; rows unchanged
	data.Sort("NONEXISTENT", false)
	if data.Rows[0][0] != "zebra" {
		t.Errorf("expected rows unchanged, got first row %q", data.Rows[0][0])
	}
}

func TestTableDataSort_CaseInsensitiveColumn(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"zebra", "Running"},
			{"apple", "Pending"},
		},
	}

	data.Sort("name", false)
	if data.Rows[0][0] != "apple" {
		t.Errorf("expected apple first, got %q", data.Rows[0][0])
	}
}

func TestTableDataSort_EmptyRows(t *testing.T) {
	data := &TableData{
		Headers: []string{"NAME"},
		Rows:    [][]string{},
	}
	// Should not panic
	data.Sort("NAME", false)
}

package list

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	headers := []string{"NAME", "STATUS"}
	rows := [][]string{
		{"pod-1", "Running"},
		{"pod-2", "Pending"},
	}

	list := New(headers, rows)

	if list.RowCount() != 2 {
		t.Errorf("RowCount() = %d, want 2", list.RowCount())
	}

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", list.SelectedIndex())
	}
}

func TestSetItems(t *testing.T) {
	list := New([]string{"A"}, [][]string{{"1"}, {"2"}, {"3"}})
	list.MoveDown()
	list.MoveDown()

	// SetItems should reset cursor
	list.SetItems([]string{"B"}, [][]string{{"x"}, {"y"}})

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() after SetItems = %d, want 0", list.SelectedIndex())
	}

	if list.RowCount() != 2 {
		t.Errorf("RowCount() after SetItems = %d, want 2", list.RowCount())
	}
}

func TestMoveUp(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})
	list.MoveDown()
	list.MoveDown()

	list.MoveUp()
	if list.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() after MoveUp = %d, want 1", list.SelectedIndex())
	}

	// Move up at top should stay at 0
	list.MoveToTop()
	list.MoveUp()
	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() at top after MoveUp = %d, want 0", list.SelectedIndex())
	}
}

func TestMoveDown(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	list.MoveDown()
	if list.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() after MoveDown = %d, want 1", list.SelectedIndex())
	}

	// Move down at bottom should stay at last
	list.MoveToBottom()
	list.MoveDown()
	if list.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() at bottom after MoveDown = %d, want 2", list.SelectedIndex())
	}
}

func TestMoveToTop(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})
	list.MoveDown()
	list.MoveDown()

	list.MoveToTop()
	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() after MoveToTop = %d, want 0", list.SelectedIndex())
	}
}

func TestMoveToBottom(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	list.MoveToBottom()
	if list.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() after MoveToBottom = %d, want 2", list.SelectedIndex())
	}
}

func TestSelectedItem(t *testing.T) {
	rows := [][]string{
		{"pod-1", "Running"},
		{"pod-2", "Pending"},
	}
	list := New([]string{"NAME", "STATUS"}, rows)

	item := list.SelectedItem()
	if item[0] != "pod-1" {
		t.Errorf("SelectedItem()[0] = %q, want %q", item[0], "pod-1")
	}

	list.MoveDown()
	item = list.SelectedItem()
	if item[0] != "pod-2" {
		t.Errorf("SelectedItem()[0] after MoveDown = %q, want %q", item[0], "pod-2")
	}
}

func TestSelectedItemEmpty(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{})

	item := list.SelectedItem()
	if item != nil {
		t.Errorf("SelectedItem() on empty list = %v, want nil", item)
	}
}

func TestHeaders(t *testing.T) {
	headers := []string{"NAME", "STATUS", "AGE"}
	list := New(headers, [][]string{})

	got := list.Headers()
	if len(got) != len(headers) {
		t.Errorf("Headers() length = %d, want %d", len(got), len(headers))
	}

	for i, h := range headers {
		if got[i] != h {
			t.Errorf("Headers()[%d] = %q, want %q", i, got[i], h)
		}
	}
}

func TestViewContainsHeaders(t *testing.T) {
	list := New([]string{"NAME", "STATUS"}, [][]string{{"pod-1", "Running"}})
	list.SetSize(80, 10)
	view := list.View()

	if !strings.Contains(view, "NAME") {
		t.Error("View should contain 'NAME' header")
	}

	if !strings.Contains(view, "STATUS") {
		t.Error("View should contain 'STATUS' header")
	}
}

func TestViewContainsRows(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"pod-1"}, {"pod-2"}})
	list.SetSize(80, 10)
	view := list.View()

	if !strings.Contains(view, "pod-1") {
		t.Error("View should contain 'pod-1'")
	}

	if !strings.Contains(view, "pod-2") {
		t.Error("View should contain 'pod-2'")
	}
}

func TestViewContainsCursor(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"pod-1"}, {"pod-2"}})
	list.SetSize(80, 10)
	view := list.View()

	if !strings.Contains(view, ">") {
		t.Error("View should contain cursor indicator '>'")
	}
}

func TestViewEmptyData(t *testing.T) {
	list := New([]string{}, [][]string{})
	list.SetSize(80, 10)
	view := list.View()

	if !strings.Contains(view, "No data") {
		t.Error("View with no headers should show 'No data'")
	}
}

func TestViewportScrolling(t *testing.T) {
	rows := [][]string{
		{"a"}, {"b"}, {"c"}, {"d"}, {"e"},
		{"f"}, {"g"}, {"h"}, {"i"}, {"j"},
	}
	list := New([]string{"NAME"}, rows)
	list.SetSize(80, 4) // 3 visible rows (1 for header)

	// Initially should show a, b, c
	view := list.View()
	if !strings.Contains(view, "a") {
		t.Error("Initial view should contain 'a'")
	}

	// Move to bottom
	list.MoveToBottom()
	view = list.View()

	// Should now show h, i, j (last 3)
	if !strings.Contains(view, "j") {
		t.Error("View after MoveToBottom should contain 'j'")
	}
}

func TestString(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}})
	s := list.String()

	if !strings.Contains(s, "rows=2") {
		t.Errorf("String() = %q, should contain 'rows=2'", s)
	}

	if !strings.Contains(s, "cursor=0") {
		t.Errorf("String() = %q, should contain 'cursor=0'", s)
	}
}

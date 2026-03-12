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

// Multi-select tests

func TestToggleSelect(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	// Initially nothing selected
	if list.SelectedCount() != 0 {
		t.Errorf("Initial SelectedCount() = %d, want 0", list.SelectedCount())
	}

	// Toggle select first item
	list.ToggleSelect()
	if !list.IsSelected(0) {
		t.Error("Item 0 should be selected after ToggleSelect")
	}
	if list.SelectedCount() != 1 {
		t.Errorf("SelectedCount() after first toggle = %d, want 1", list.SelectedCount())
	}

	// Toggle same item again should deselect
	list.ToggleSelect()
	if list.IsSelected(0) {
		t.Error("Item 0 should be deselected after second ToggleSelect")
	}
	if list.SelectedCount() != 0 {
		t.Errorf("SelectedCount() after deselect = %d, want 0", list.SelectedCount())
	}
}

func TestToggleSelectMultiple(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	// Select first item
	list.ToggleSelect()

	// Move down and select second item
	list.MoveDown()
	list.ToggleSelect()

	if list.SelectedCount() != 2 {
		t.Errorf("SelectedCount() = %d, want 2", list.SelectedCount())
	}

	if !list.IsSelected(0) {
		t.Error("Item 0 should be selected")
	}

	if !list.IsSelected(1) {
		t.Error("Item 1 should be selected")
	}

	if list.IsSelected(2) {
		t.Error("Item 2 should not be selected")
	}
}

func TestSelectAll(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	list.SelectAll()

	if list.SelectedCount() != 3 {
		t.Errorf("SelectedCount() after SelectAll = %d, want 3", list.SelectedCount())
	}

	for i := 0; i < 3; i++ {
		if !list.IsSelected(i) {
			t.Errorf("Item %d should be selected after SelectAll", i)
		}
	}
}

func TestDeselectAll(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	list.SelectAll()
	list.DeselectAll()

	if list.SelectedCount() != 0 {
		t.Errorf("SelectedCount() after DeselectAll = %d, want 0", list.SelectedCount())
	}

	for i := 0; i < 3; i++ {
		if list.IsSelected(i) {
			t.Errorf("Item %d should not be selected after DeselectAll", i)
		}
	}
}

func TestSelectedItems(t *testing.T) {
	rows := [][]string{
		{"pod-1", "Running"},
		{"pod-2", "Pending"},
		{"pod-3", "Failed"},
	}
	list := New([]string{"NAME", "STATUS"}, rows)

	// With no selections, should return current item
	items := list.SelectedItems()
	if len(items) != 1 {
		t.Errorf("SelectedItems() with no selection = %d items, want 1", len(items))
	}
	if items[0][0] != "pod-1" {
		t.Errorf("SelectedItems()[0][0] = %q, want %q", items[0][0], "pod-1")
	}

	// Select specific items
	list.ToggleSelect() // select pod-1
	list.MoveDown()
	list.MoveDown()
	list.ToggleSelect() // select pod-3

	items = list.SelectedItems()
	if len(items) != 2 {
		t.Errorf("SelectedItems() with 2 selections = %d items, want 2", len(items))
	}

	// Items should be in order (0, 2)
	if items[0][0] != "pod-1" {
		t.Errorf("SelectedItems()[0][0] = %q, want %q", items[0][0], "pod-1")
	}
	if items[1][0] != "pod-3" {
		t.Errorf("SelectedItems()[1][0] = %q, want %q", items[1][0], "pod-3")
	}
}

func TestSelectedItemsEmpty(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{})

	items := list.SelectedItems()
	if items != nil {
		t.Errorf("SelectedItems() on empty list = %v, want nil", items)
	}
}

func TestIsSelected(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})

	// Initially none selected
	for i := 0; i < 3; i++ {
		if list.IsSelected(i) {
			t.Errorf("IsSelected(%d) = true, want false initially", i)
		}
	}

	// Select item 1
	list.MoveDown()
	list.ToggleSelect()

	if list.IsSelected(0) {
		t.Error("IsSelected(0) = true, want false")
	}
	if !list.IsSelected(1) {
		t.Error("IsSelected(1) = false, want true")
	}
	if list.IsSelected(2) {
		t.Error("IsSelected(2) = true, want false")
	}
}

func TestViewContainsSelectionIndicator(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"pod-1"}, {"pod-2"}})
	list.SetSize(80, 10)

	// Select first item
	list.ToggleSelect()
	view := list.View()

	// Should show selection indicator
	if !strings.Contains(view, "*") {
		t.Error("View should contain '*' selection indicator for marked item")
	}
}

func TestSetItemsClearsSelection(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}})

	list.SelectAll()
	if list.SelectedCount() != 2 {
		t.Errorf("SelectedCount() after SelectAll = %d, want 2", list.SelectedCount())
	}

	// SetItems should clear selections
	list.SetItems([]string{"NAME"}, [][]string{{"x"}, {"y"}, {"z"}})

	if list.SelectedCount() != 0 {
		t.Errorf("SelectedCount() after SetItems = %d, want 0", list.SelectedCount())
	}
}

func TestUpdateItemsPreservesCursor(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})
	list.MoveDown()
	list.MoveDown() // cursor at 2

	// UpdateItems should preserve cursor position
	list.UpdateItems([]string{"NAME"}, [][]string{{"x"}, {"y"}, {"z"}})

	if list.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() after UpdateItems = %d, want 2", list.SelectedIndex())
	}
	if list.RowCount() != 3 {
		t.Errorf("RowCount() after UpdateItems = %d, want 3", list.RowCount())
	}
}

func TestUpdateItemsClampsCursor(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}, {"d"}})
	list.MoveToBottom() // cursor at 3

	// Shrink the list — cursor should clamp to last item
	list.UpdateItems([]string{"NAME"}, [][]string{{"x"}, {"y"}})

	if list.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() after UpdateItems with fewer rows = %d, want 1", list.SelectedIndex())
	}
}

func TestUpdateItemsClampsCursorToZeroOnEmpty(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}})
	list.MoveDown() // cursor at 1

	list.UpdateItems([]string{"NAME"}, [][]string{})

	if list.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() after UpdateItems with empty rows = %d, want 0", list.SelectedIndex())
	}
}

func TestUpdateItemsPrunesOutOfRangeSelections(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}, {"c"}})
	list.SelectAll() // 0, 1, 2 selected

	// Shrink to 2 rows — selection at index 2 should be pruned
	list.UpdateItems([]string{"NAME"}, [][]string{{"x"}, {"y"}})

	if list.SelectedCount() != 2 {
		t.Errorf("SelectedCount() after UpdateItems = %d, want 2", list.SelectedCount())
	}
	if !list.IsSelected(0) {
		t.Error("Item 0 should still be selected")
	}
	if !list.IsSelected(1) {
		t.Error("Item 1 should still be selected")
	}
	if list.IsSelected(2) {
		t.Error("Item 2 should have been pruned")
	}
}

func TestStringContainsSelectedCount(t *testing.T) {
	list := New([]string{"NAME"}, [][]string{{"a"}, {"b"}})
	list.SelectAll()

	s := list.String()
	if !strings.Contains(s, "selected=2") {
		t.Errorf("String() = %q, should contain 'selected=2'", s)
	}
}

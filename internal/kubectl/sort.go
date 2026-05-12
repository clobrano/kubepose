package kubectl

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// SortByCreationTime is the special sort_by value that sorts by the AGE column.
const SortByCreationTime = "creation_time"

// Sort sorts the table rows by the given column name (case-insensitive) or by
// SortByCreationTime (which uses the AGE column).  When reverse is true the
// order is reversed.  The headers are never modified.  If the requested column
// cannot be found the rows are left in their original order.
func (t *TableData) Sort(by string, reverse bool) {
	if len(t.Rows) == 0 {
		return
	}

	colIdx := t.resolveColumnIndex(by)
	if colIdx < 0 {
		return
	}

	useAge := strings.EqualFold(by, SortByCreationTime) || strings.EqualFold(t.Headers[colIdx], "AGE")

	sort.SliceStable(t.Rows, func(i, j int) bool {
		a := cellValue(t.Rows[i], colIdx)
		b := cellValue(t.Rows[j], colIdx)

		var less bool
		if useAge {
			// Shorter age = more recently created; sort ascending = newest first.
			less = parseAge(a) < parseAge(b)
		} else {
			less = a < b
		}

		if reverse {
			return !less
		}
		return less
	})
}

// resolveColumnIndex returns the column index for the given sort key.
// "creation_time" maps to the AGE column.
func (t *TableData) resolveColumnIndex(by string) int {
	if strings.EqualFold(by, SortByCreationTime) {
		return t.GetColumnIndex("AGE")
	}
	return t.GetColumnIndex(by)
}

// cellValue returns the value at colIdx in row, or "" when out of bounds.
func cellValue(row []string, colIdx int) string {
	if colIdx < len(row) {
		return row[colIdx]
	}
	return ""
}

// parseAge converts a kubectl age string (e.g. "5s", "2m", "1h30m", "3d4h")
// to a time.Duration.  Unknown / unparseable values map to 0.
func parseAge(age string) time.Duration {
	age = strings.TrimSpace(age)
	if age == "" || age == "<unknown>" || age == "-" {
		return 0
	}

	var total time.Duration
	// Walk through the string collecting digit runs followed by a unit char.
	s := age
	for len(s) > 0 {
		// Consume digits
		numEnd := 0
		for numEnd < len(s) && unicode.IsDigit(rune(s[numEnd])) {
			numEnd++
		}
		if numEnd == 0 || numEnd >= len(s) {
			break
		}
		n, err := strconv.Atoi(s[:numEnd])
		if err != nil {
			break
		}
		unit := s[numEnd]
		s = s[numEnd+1:]

		switch unit {
		case 'd':
			total += time.Duration(n) * 24 * time.Hour
		case 'h':
			total += time.Duration(n) * time.Hour
		case 'm':
			total += time.Duration(n) * time.Minute
		case 's':
			total += time.Duration(n) * time.Second
		default:
			// Unknown unit – skip
		}
	}
	return total
}

package search

import (
	"strings"
	"unicode"
)

// FuzzyMatch checks if pattern fuzzy-matches text and returns match status and score
// Characters in pattern must appear in order in text, but not necessarily adjacent
// Higher score = better match (exact match scores highest)
func FuzzyMatch(pattern, text string) (bool, int) {
	if pattern == "" {
		return true, 0
	}

	// Case-insensitive matching
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	patternIdx := 0
	score := 0
	lastMatchIdx := -1
	consecutiveBonus := 0

	for i, ch := range text {
		if patternIdx >= len(pattern) {
			break
		}

		patternCh := rune(pattern[patternIdx])
		if ch == patternCh {
			// Match found
			patternIdx++
			score += 10 // Base score for match

			// Bonus for consecutive matches
			if lastMatchIdx == i-1 {
				consecutiveBonus++
				score += consecutiveBonus * 5
			} else {
				consecutiveBonus = 0
			}

			// Bonus for match at start of word
			if i == 0 || !unicode.IsLetter(rune(text[i-1])) {
				score += 15
			}

			// Bonus for match at start of text
			if i == 0 {
				score += 20
			}

			lastMatchIdx = i
		}
	}

	// All pattern characters must be matched
	if patternIdx < len(pattern) {
		return false, 0
	}

	// Penalty for longer text (prefer shorter matches)
	score -= len(text) / 10

	// Bonus for exact match
	if pattern == text {
		score += 100
	}

	return true, score
}

// FuzzyMatchRow checks if pattern matches any cell in a row
// Returns match status and best score across all cells
func FuzzyMatchRow(pattern string, row []string) (bool, int) {
	if pattern == "" {
		return true, 0
	}

	bestScore := 0
	matched := false

	for _, cell := range row {
		if match, score := FuzzyMatch(pattern, cell); match {
			matched = true
			if score > bestScore {
				bestScore = score
			}
		}
	}

	return matched, bestScore
}

// FilterRows filters rows based on fuzzy matching against any column
// Returns filtered rows sorted by match score (best first)
func FilterRows(pattern string, rows [][]string) [][]string {
	if pattern == "" {
		return rows
	}

	type scoredRow struct {
		row   []string
		score int
	}

	var scored []scoredRow

	for _, row := range rows {
		if match, score := FuzzyMatchRow(pattern, row); match {
			scored = append(scored, scoredRow{row: row, score: score})
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	result := make([][]string, len(scored))
	for i, s := range scored {
		result[i] = s.row
	}

	return result
}

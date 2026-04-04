package chunk

import (
	"strings"
	"unicode"
)

func IsHighValueHeading(text string, fontSize float64) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	if len(text) < 5 {
		return false
	}
	// Rule 1: Physical certainty (Font Size is king)
	if fontSize > 13.5 {
		return true
	}

	// Rule 2: Fallback to your existing casing logic for standard fonts
	runes := []rune(text)
	upper, letter := 0, 0
	for _, r := range runes {
		if unicode.IsLetter(r) {
			letter++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	// If it's short and mostly uppercase, it's a structural label
	if letter > 0 && upper == letter && len(text) < 40 {
		return true
	}

	return false
}

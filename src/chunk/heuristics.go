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
	if fontSize > 13.5 {
		return true
	}
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
	if letter > 0 && upper == letter && len(text) < 40 {
		return true
	}

	return false
}

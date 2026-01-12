package chunk

import (
	"regexp"
	"strings"
	"unicode"
)

var numericHeading = regexp.MustCompile(`^\d+(\.\d+)*\s+\S+`)

func LooksLikeHeading(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	runes := []rune(s)
	length := len(runes)
	if length > 120 {
		return false
	}
	if length < 80 && strings.HasSuffix(s, ":") {
		return true
	}
	if numericHeading.MatchString(s) {
		return true
	}

	upper := 0
	letter := 0

	for _, r := range runes {
		if unicode.IsLetter(r) {
			letter++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}

	if letter > 0 && upper*2 >= letter && length < 60 {
		return true
	}

	return false
}

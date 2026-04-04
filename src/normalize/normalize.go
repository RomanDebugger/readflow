package normalize

import (
	"strings"
	"unicode"
)

func NormalizeText(s string) string {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)

	return strings.Join(strings.Fields(clean), " ")
}

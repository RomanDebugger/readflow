package chunk

import "strings"

func LooksLikeHeading(s string) bool {
	if len(s) < 80 && strings.HasSuffix(s, ":") {
		return true
	}

	upper := 0
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			upper++
		}
	}

	return len(s) < 60 && upper > len(s)/2
}

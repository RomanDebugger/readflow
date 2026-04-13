package chunk

import (
	"strings"
	"unicode"
)

func ScoreChunk(c *Chunk) float32 {
	var score float32 = 0.5

	if c.Type == "title" {
		return 0.95
	}

	numCount := 0
	for _, r := range c.Text {
		if unicode.IsDigit(r) {
			numCount++
		}
	}

	if numCount > 5 {
		score += 0.2
	}

	if strings.Count(c.Text, ".") > 12 || strings.Count(c.Text, "---") > 2 {
		score -= 0.3
	}

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.1 {
		score = 0.1
	}
	return score
}

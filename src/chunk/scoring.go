package chunk

import (
	"strings"
)

func ScoreChunk(c *Chunk) float32 {
	if c.Type == "title" {
		return 0.95
	} // Titles are high value

	var score float32 = 0.5

	// Bonus for "Content Density" (Word length average)
	words := strings.Fields(c.Text)
	if len(words) > 0 {
		avgLen := 0
		for _, w := range words {
			avgLen += len(w)
		}
		if (avgLen / len(words)) > 5 {
			score += 0.2
		} // Complex words = Higher value
	}

	// Penalty for "Artifacts" (Too many symbols/dots)
	if strings.Count(c.Text, ".") > 10 || strings.Count(c.Text, "-") > 10 {
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

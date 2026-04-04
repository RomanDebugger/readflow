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

	// 1. Technical Density Check
	// If a chunk contains a mix of letters, numbers, and units (kg, V, MHz),
	// it's high-signal technical data.
	numCount := 0
	for _, r := range c.Text {
		if unicode.IsDigit(r) {
			numCount++
		}
	}

	if numCount > 5 {
		score += 0.2 // It's likely a specification or data point
	}

	// 2. Artifact Penalty (Dots/Dashes/Messy Extraction)
	// If the Go engine sees too much punctuation, it's likely a TOC or a messy table
	if strings.Count(c.Text, ".") > 12 || strings.Count(c.Text, "---") > 2 {
		score -= 0.3
	}

	// 3. Format Normalization
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.1 {
		score = 0.1
	}
	return score
}

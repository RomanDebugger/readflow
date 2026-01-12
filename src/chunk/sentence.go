package chunk

import (
	"regexp"
	"strings"
)

var sentenceEnd = regexp.MustCompile(`([.!?])\s+`)

func SplitIntoSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	for i := 0; i < len(text); i++ {
		current.WriteByte(text[i])

		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			if i+1 == len(text) || text[i+1] == ' ' || text[i+1] == '\n' {
				s := strings.TrimSpace(current.String())
				if len(s) > 0 {
					sentences = append(sentences, s)
				}
				current.Reset()
			}
		}
	}

	rest := strings.TrimSpace(current.String())
	if len(rest) > 0 {
		sentences = append(sentences, rest)
	}

	return sentences
}

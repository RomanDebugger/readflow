package chunk

import (
	"strings"
)

func SplitIntoSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	for i := 0; i < len(text); i++ {
		current.WriteByte(text[i])

		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
			continue
		}

		if current.Len() >= 200 {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}

	rest := strings.TrimSpace(current.String())
	if rest != "" {
		sentences = append(sentences, rest)
	}

	return sentences
}

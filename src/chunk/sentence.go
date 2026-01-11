package chunk

import (
	"regexp"
	"strings"
)

var sentenceEnd = regexp.MustCompile(`([.!?])\s+`)

func SplitIntoSentences(text string) []string {
	raw := sentenceEnd.Split(text, -1)
	var sentences []string

	for _, s := range raw {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

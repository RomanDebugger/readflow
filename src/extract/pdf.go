package extract

import (
	"sort"
	"strings"

	"rsc.io/pdf"
)

type TextChunk struct {
	Text     string  `json:"text"`
	Y        float64 `json:"y"`
	X        float64 `json:"x"`
	FontSize float64 `json:"font_size"`
	Type     string  `json:"type"`
}

type PageData struct {
	Page   int         `json:"page"`
	Chunks []TextChunk `json:"chunks"`
	// Keep a flat text version for legacy compatibility if needed
	RawText string `json:"raw_text"`
}

type DocumentText struct {
	Document string     `json:"document"`
	Pages    []PageData `json:"pages"`
}

func safePageContent(p pdf.Page) (content pdf.Content, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	content = p.Content()
	return content, true
}
func ExtractText(path string) (*DocumentText, error) {
	doc := DocumentText{
		Document: path,
		Pages:    []PageData{},
	}

	r, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		var height float64 = 842
		box := p.V.Key("MediaBox")
		if box.Kind() == pdf.Array && box.Len() == 4 {
			height = box.Index(3).Float64()
		}

		content, ok := safePageContent(p)
		if !ok {
			continue
		}

		var chunks []TextChunk
		var currentWord strings.Builder
		var lastY, lastX, lastFontSize float64
		var fullPageText strings.Builder

		// Helper to determine type and append
		flush := func(text string, y, x, fSize float64) {
			cleanText := strings.TrimSpace(text)
			if cleanText == "" {
				return
			}

			cType := "body"
			if y > (height * 0.92) {
				cType = "header"
			} else if y < (height * 0.08) {
				cType = "footer"
			} else if fSize > 13 {
				cType = "heading"
			}

			chunks = append(chunks, TextChunk{
				Text:     cleanText,
				Y:        y,
				X:        x,
				FontSize: fSize,
				Type:     cType,
			})
			fullPageText.WriteString(cleanText + " ")
		}

		for _, txt := range content.Text {
			// SPATIAL RECOMPOSITION LOGIC
			// If the character is on the same line and "close enough" horizontally
			// We treat 0.25 * FontSize as a reasonable threshold for "same word"
			isSameLine := txt.Y == lastY
			isClose := (txt.X - lastX) < (txt.FontSize * 0.5)

			if isSameLine && isClose {
				currentWord.WriteString(txt.S)
			} else {
				// Gap detected! Flush the finished word/phrase
				flush(currentWord.String(), lastY, lastX, lastFontSize)
				currentWord.Reset()
				currentWord.WriteString(txt.S)
			}

			lastY = txt.Y
			// We update lastX to be the END of the current character
			lastX = txt.X + (float64(len(txt.S)) * txt.FontSize * 0.45)
			lastFontSize = txt.FontSize
		}

		// Final flush for the last word on the page
		flush(currentWord.String(), lastY, lastX, lastFontSize)

		// Sort chunks by Y (Top to Bottom) and X (Left to Right)
		sort.Slice(chunks, func(i, j int) bool {
			if chunks[i].Y != chunks[j].Y {
				return chunks[i].Y > chunks[j].Y
			}
			return chunks[i].X < chunks[j].X
		})

		doc.Pages = append(doc.Pages, PageData{
			Page:    i,
			Chunks:  chunks,
			RawText: fullPageText.String(),
		})
	}

	return &doc, nil
}

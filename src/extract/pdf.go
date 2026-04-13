package extract

import (
	"math"
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
	Page    int         `json:"page"`
	Chunks  []TextChunk `json:"chunks"`
	RawText string      `json:"raw_text"`
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
			yTolerance := txt.FontSize * 0.3
			isSameLine := math.Abs(txt.Y-lastY) < yTolerance
			gap := txt.X - lastX

			if isSameLine {
				if gap < (txt.FontSize*0.2) || gap < 0 {
					currentWord.WriteString(txt.S)
				} else if gap < (txt.FontSize * 1.5) {
					currentWord.WriteString(" ")
					currentWord.WriteString(txt.S)
				} else {
					flush(currentWord.String(), lastY, lastX, lastFontSize)
					currentWord.Reset()
					currentWord.WriteString(txt.S)
				}
			} else {
				if currentWord.Len() > 0 {
					flush(currentWord.String(), lastY, lastX, lastFontSize)
				}
				currentWord.Reset()
				currentWord.WriteString(txt.S)
			}

			lastY = txt.Y
			lastX = txt.X + (float64(len(txt.S)) * txt.FontSize * 0.5)
			lastFontSize = txt.FontSize
		}
		flush(currentWord.String(), lastY, lastX, lastFontSize)

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

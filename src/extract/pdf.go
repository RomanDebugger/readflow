package extract

import (
	"strings"

	"rsc.io/pdf"
)

type PageText struct {
	Page int    `json:"page"`
	Text string `json:"text"`
}

type DocumentText struct {
	Document string     `json:"document"`
	Pages    []PageText `json:"pages"`
}

func ExtractText(path string) (*DocumentText, error) {

	doc := DocumentText{
		Document: path,
		Pages:    []PageText{},
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

		var pageText strings.Builder

		content := p.Content()
		for _, txt := range content.Text {
			pageText.WriteString(txt.S)
		}

		doc.Pages = append(doc.Pages, PageText{
			Page: i,
			Text: pageText.String(),
		})
	}

	return &doc, nil
}

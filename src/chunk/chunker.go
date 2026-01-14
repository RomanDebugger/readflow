package chunk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"readflow/src/extract"
)

const MaxChunkSize = 600

func ChunkDocument(doc extract.DocumentText, outDir string) error {
	var chunks []Chunk
	chunkIndex := 0

	for _, page := range doc.Pages {
		sentences := SplitIntoSentences(page.Text)

		var buffer strings.Builder

		flush := func() {
			text := strings.TrimSpace(buffer.String())
			if text == "" {
				buffer.Reset()
				return
			}

			chunks = append(chunks, Chunk{
				ChunkID: fmt.Sprintf("%s_p%d_c%d", doc.Document, page.Page, chunkIndex),
				DocID:   doc.Document,
				Page:    page.Page,
				Index:   chunkIndex,
				Text:    text,
				Length:  len(text),
			})

			chunkIndex++
			buffer.Reset()
		}

		for _, sentence := range sentences {
			// Defensive cap: trim pathological sentences
			if len(sentence) > MaxChunkSize {
				sentence = sentence[:MaxChunkSize]
			}

			// If adding this sentence would overflow, flush first
			if buffer.Len()+len(sentence)+1 > MaxChunkSize {
				flush()
			}

			if buffer.Len() > 0 {
				buffer.WriteByte(' ')
			}
			buffer.WriteString(sentence)
		}

		flush()
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	base := filepath.Base(doc.Document)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	outPath := filepath.Join(outDir, name+".json")

	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(chunks)
}

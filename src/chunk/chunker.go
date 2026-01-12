package chunk

import (
	"encoding/json"
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
		charStart := 0
		// NOTE: CharStart/CharEnd are page-local offsets in v1

		for _, sentence := range sentences {
			if LooksLikeHeading(sentence) && buffer.Len() > 0 {
				chunks = append(chunks, Chunk{
					DocID:      doc.Document,
					Page:       page.Page,
					ChunkIndex: chunkIndex,
					CharStart:  charStart,
					CharEnd:    charStart + buffer.Len(),
					Text:       buffer.String(),
				})
				chunkIndex++
				buffer.Reset()
			}

			if buffer.Len()+len(sentence) > MaxChunkSize {
				chunks = append(chunks, Chunk{
					DocID:      doc.Document,
					Page:       page.Page,
					ChunkIndex: chunkIndex,
					CharStart:  charStart,
					CharEnd:    charStart + buffer.Len(),
					Text:       buffer.String(),
				})
				chunkIndex++
				buffer.Reset()
			}

			if buffer.Len() == 0 {
				charStart = 0
			}
			// TODO: charStart currently resets per chunk (v1 approximation)

			buffer.WriteString(sentence)
			buffer.WriteString(" ")
		}

		if buffer.Len() > 0 {
			chunks = append(chunks, Chunk{
				DocID:      doc.Document,
				Page:       page.Page,
				ChunkIndex: chunkIndex,
				CharStart:  charStart,
				CharEnd:    charStart + buffer.Len(),
				Text:       buffer.String(),
			})
			chunkIndex++
		}
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

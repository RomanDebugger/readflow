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

		// Helper to create and score a chunk
		createChunk := func(t string, cType string) {
			text := strings.TrimSpace(t)
			if text == "" {
				return
			}

			newChunk := Chunk{
				ChunkID: fmt.Sprintf("%s_p%d_c%d", filepath.Base(doc.Document), page.Page, chunkIndex),
				DocID:   doc.Document,
				Page:    page.Page,
				Index:   chunkIndex,
				Text:    text,
				Length:  len(text),
				Type:    cType,
			}

			// Apply our new "Simpler" Scorer
			newChunk.Quality = ScoreChunk(&newChunk)

			chunks = append(chunks, newChunk)
			chunkIndex++
		}

		for _, sentence := range sentences {
			isHeading := LooksLikeHeading(sentence)

			// IF it's a heading: Flush existing buffer as paragraph,
			// then flush heading as its own title chunk
			if isHeading {
				if buffer.Len() > 0 {
					createChunk(buffer.String(), "paragraph")
					buffer.Reset()
				}
				createChunk(sentence, "title")
				continue
			}

			// Normal Chunking Logic
			if buffer.Len()+len(sentence)+1 > MaxChunkSize {
				createChunk(buffer.String(), "paragraph")
				buffer.Reset()
			}

			if buffer.Len() > 0 {
				buffer.WriteByte(' ')
			}
			buffer.WriteString(sentence)
		}

		// Final flush for the page
		if buffer.Len() > 0 {
			createChunk(buffer.String(), "paragraph")
			buffer.Reset()
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

package chunk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"readflow/src/extract"
)

const MaxChunkSize = 750

func ChunkDocument(doc extract.DocumentText, outDir string) error {
	var chunks []Chunk
	var buffer strings.Builder
	chunkIndex := 0

	for _, page := range doc.Pages {

		for _, raw := range page.Chunks {
			if raw.Type == "header" || raw.Type == "footer" {
				continue
			}
			isHeading := raw.Type == "heading" || IsHighValueHeading(raw.Text, raw.FontSize)

			if isHeading {
				if buffer.Len() > 0 {
					chunks = append(chunks, finalizeChunk(buffer.String(), "paragraph", page.Page, chunkIndex, doc.Document, raw.FontSize))
					buffer.Reset()
					chunkIndex++
				}
				chunks = append(chunks, finalizeChunk(raw.Text, "title", page.Page, chunkIndex, doc.Document, raw.FontSize))
				chunkIndex++
				continue
			}
			sentences := SplitIntoSentences(raw.Text)

			for _, s := range sentences {
				if buffer.Len()+len(s)+1 > MaxChunkSize {
					chunks = append(chunks, finalizeChunk(buffer.String(), "paragraph", page.Page, chunkIndex, doc.Document, raw.FontSize))
					buffer.Reset()
					chunkIndex++
				}

				if buffer.Len() > 0 {
					buffer.WriteByte(' ')
				}
				buffer.WriteString(s)
			}
		}
	}
	if buffer.Len() > 0 {
		chunks = append(chunks, finalizeChunk(buffer.String(), "paragraph", 0, chunkIndex, doc.Document, 0))
	}

	return saveToFile(chunks, doc.Document, outDir)
}

func finalizeChunk(text string, cType string, page int, idx int, docPath string, fSize float64) Chunk {
	c := Chunk{
		ChunkID: fmt.Sprintf("%s_c%d", filepath.Base(docPath), idx),
		DocID:   docPath,
		Page:    page,
		Index:   idx,
		Text:    strings.TrimSpace(text),
		Length:  len(text),
		Type:    cType,
		Metadata: Metadata{
			SourceFile: docPath,
			FontSize:   fSize,
		},
	}

	c.Quality = ScoreChunk(&c)
	return c
}

func saveToFile(chunks []Chunk, docPath string, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	base := filepath.Base(docPath)
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

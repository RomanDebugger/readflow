package chunk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"readflow/src/extract"
)

const MaxChunkSize = 750 // Increased slightly for better Gemini reasoning

func ChunkDocument(doc extract.DocumentText, outDir string) error {
	var chunks []Chunk
	var buffer strings.Builder
	chunkIndex := 0

	for _, page := range doc.Pages {
		// Reset buffer logic per page if you want to avoid merging page boundaries,
		// but usually, continuing across pages is better for flow.

		for _, raw := range page.Chunks {

			// 1. THE "NOISE GATE": Drop headers/footers identified by Go
			if raw.Type == "header" || raw.Type == "footer" {
				continue
			}

			// 2. HEADING DETECTION: Uses the new Font-Size heuristic
			// We check the raw chunk type AND your existing heuristic logic
			isHeading := raw.Type == "heading" || IsHighValueHeading(raw.Text, raw.FontSize)

			if isHeading {
				// Flush current buffer as a paragraph before starting the new section
				if buffer.Len() > 0 {
					chunks = append(chunks, finalizeChunk(buffer.String(), "paragraph", page.Page, chunkIndex, doc.Document, raw.FontSize))
					buffer.Reset()
					chunkIndex++
				}
				// Create the Heading/Title chunk immediately
				chunks = append(chunks, finalizeChunk(raw.Text, "title", page.Page, chunkIndex, doc.Document, raw.FontSize))
				chunkIndex++
				continue
			}

			// 3. SMART SENTENCE SPLITTING:
			// Instead of raw text, we split the individual body chunk into sentences
			sentences := SplitIntoSentences(raw.Text)

			for _, s := range sentences {
				// If adding this sentence exceeds MaxChunkSize, flush buffer
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

	// Final Flush for any remaining text
	if buffer.Len() > 0 {
		chunks = append(chunks, finalizeChunk(buffer.String(), "paragraph", 0, chunkIndex, doc.Document, 0))
	}

	return saveToFile(chunks, doc.Document, outDir)
}

// finalizeChunk handles the creation, metadata attachment, and scoring
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

	// Apply your upgraded Scoring logic
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

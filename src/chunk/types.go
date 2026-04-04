package chunk

type Metadata struct {
	SourceFile string  `json:"source"`
	FontSize   float64 `json:"font_size"`
}

type Chunk struct {
	ChunkID string `json:"chunk_id"`
	DocID   string `json:"doc_id"`
	Page    int    `json:"page"`
	Index   int    `json:"index"`

	Text   string `json:"text"`
	Length int    `json:"length"`

	Type    string  `json:"type"`    // "title", "paragraph", "spec"
	Quality float32 `json:"quality"` // 0.0 – 1.0

	Metadata Metadata `json:"metadata"`
}

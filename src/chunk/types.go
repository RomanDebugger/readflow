package chunk

// CharStart and CharEnd are relative to the page text, not the full document

type Chunk struct {
	DocID      string `json:"doc_id"`
	Page       int    `json:"page"`
	ChunkIndex int    `json:"chunk_index"`
	CharStart  int    `json:"char_start"`
	CharEnd    int    `json:"char_end"`
	Text       string `json:"text"`
}

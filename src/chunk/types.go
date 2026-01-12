package chunk

// CharStart and CharEnd are relative to the page text, not the full document

type Chunk struct {
	DocID      string `json:"doc_id"`
	Page       int    `json:"page"`
	ChunkIndex int    `json:"chunk_index"`

	CharStart int `json:"char_start"`
	CharEnd   int `json:"char_end"`

	Text string `json:"text"`

	Type    string  `json:"type"`    // paragraph, title, auto_split, noise
	Quality float32 `json:"quality"` // heuristic confidence

	PrevChunk *int `json:"prev_chunk,omitempty"`
	NextChunk *int `json:"next_chunk,omitempty"`
}

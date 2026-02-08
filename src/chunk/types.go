package chunk

type Chunk struct {
	ChunkID string `json:"chunk_id"`

	DocID string `json:"doc_id"`
	Page  int    `json:"page"`
	Index int    `json:"index"`

	Text   string `json:"text"`
	Length int    `json:"length"`

	Type    string  `json:"type"`    // title, paragraph, auto_split, noise
	Quality float32 `json:"quality"` // 0.0 – 1.0

	PrevIndex *int `json:"prev_index,omitempty"`
	NextIndex *int `json:"next_index,omitempty"`
}

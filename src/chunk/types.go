package chunk

type Chunk struct {
	ChunkID string `json:"chunk_id"`

	DocID string `json:"doc_id"`
	Page  int    `json:"page"`
	Index int    `json:"index"`

	Text   string `json:"text"`
	Length int    `json:"length"`
}

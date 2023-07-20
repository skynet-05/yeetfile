package shared

const ChunkSize int = 5242880 // 5 mb

type DownloadResponse struct {
	Name   string
	ID     string
	Chunks int
	Key    string
}

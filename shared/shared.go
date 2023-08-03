package shared

const ChunkSize int = 5242880 // 5 mb

type UploadMetadata struct {
	Name       string `json:"name"`
	Chunks     int    `json:"chunks"`
	Password   string `json:"password"`
	Downloads  int    `json:"downloads"`
	Expiration string `json:"expiration"`
}

type DownloadResponse struct {
	Name   string
	ID     string
	Chunks int
	Key    string
}

type Signup struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

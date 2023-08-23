package shared

const ChunkSize int = 5242880 // 5 mb

type UploadMetadata struct {
	Name       string `json:"name"`
	Chunks     int    `json:"chunks"`
	Salt       []byte `json:"salt"`
	Downloads  int    `json:"downloads"`
	Expiration string `json:"expiration"`
}

type DownloadResponse struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Salt   []byte `json:"salt"`
	Chunks int    `json:"chunks"`
}

type Signup struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

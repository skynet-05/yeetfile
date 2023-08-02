package shared

const ChunkSize int = 5242880 // 5 mb

type DownloadResponse struct {
	Name   string
	ID     string
	Chunks int
	Key    string
}

type Signup struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

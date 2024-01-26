package shared

import (
	"fmt"
	"time"
)

const NonceSize int = 24
const KeySize int = 32
const ChunkSize int = 5242880 // 5 mb
const TotalOverhead int = 40  // secretbox overhead (16) + nonce size (24)

type UploadMetadata struct {
	Name       string `json:"name"`
	Chunks     int    `json:"chunks"`
	Size       int    `json:"size"`
	Salt       []byte `json:"salt"`
	Downloads  int    `json:"downloads"`
	Expiration string `json:"expiration"`
}

type DownloadResponse struct {
	Name       string    `json:"name"`
	ID         string    `json:"id"`
	Salt       []byte    `json:"salt"`
	Size       int       `json:"size"`
	Chunks     int       `json:"chunks"`
	Downloads  int       `json:"downloads"`
	Expiration time.Time `json:"expiration"`
}

type Signup struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Login struct {
	Identifier string `json:"email"`
	Password   string `json:"password"`
}

type SessionInfo struct {
	Meter int `json:"meter"`
}

func ReadableFileSize(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGT"[exp])
}

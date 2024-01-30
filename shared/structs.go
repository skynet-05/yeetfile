package shared

import "time"

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

type ForgotPassword struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	Email           string `json:"email"`
	Code            string `json:"code"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm-password"`
}

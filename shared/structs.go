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

type PlaintextUpload struct {
	Name       string `json:"name"`
	Salt       []byte `json:"salt"`
	Downloads  int    `json:"downloads"`
	Expiration string `json:"expiration"`
	Text       []byte `json:"text"`
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
	Identifier   string `json:"identifier"`
	LoginKeyHash string `json:"loginKeyHash"`
	ProtectedKey []byte `json:"protectedKey"`
}

type SignupResponse struct {
	Identifier string `json:"identifier"`
	Captcha    string `json:"captcha"`
	Error      string `json:"error"`
}

type VerifyAccount struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	LoginKeyHash []byte `json:"loginKeyHash"`
	ProtectedKey []byte `json:"protectedKey"`
}

type Login struct {
	Identifier   string `json:"identifier"`
	LoginKeyHash []byte `json:"loginKeyHash"`
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

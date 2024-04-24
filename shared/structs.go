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

type VaultUpload struct {
	Name         string `json:"name"`
	Length       int    `json:"length"`
	Chunks       int    `json:"chunks"`
	FolderID     string `json:"folderID"`
	ProtectedKey []byte `json:"protectedKey"`
}

type ModifyVaultFolder struct {
	Name string `json:"name"`
}

type ModifyVaultFile struct {
	Name string `json:"name"`
}

type VaultUploadResponse struct {
	ID string `json:"id"`
}

type NewFolderResponse struct {
	ID string `json:"id"`
}

type VaultItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Size         int       `json:"size"`
	Modified     time.Time `json:"modified"`
	ProtectedKey []byte    `json:"protectedKey"`
	SharedWith   int       `json:"sharedWith"`
	SharedBy     string    `json:"sharedBy"`
	LinkTag      string    `json:"linkTag"`
	CanModify    bool      `json:"canModify"`
	IsOwner      bool      `json:"isOwner"`
	RefID        string    `json:"refID"`
}

type NewVaultFolder struct {
	Name         string `json:"name"`
	ProtectedKey []byte `json:"protectedKey"`
	ParentID     string `json:"parentID"`
}

type NewPublicVaultFolder struct {
	ID           string `json:"id"`
	ProtectedKey []byte `json:"protectedKey"`
	LinkTag      string `json:"linkTag"`
}

type VaultFolder struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Modified     time.Time `json:"modified"`
	ParentID     string    `json:"parentID"`
	ProtectedKey []byte    `json:"protectedKey"`
	SharedWith   int       `json:"sharedWith"`
	SharedBy     string    `json:"sharedBy"`
	LinkTag      string    `json:"linkTag"`
	CanModify    bool      `json:"canModify"`
	RefID        string    `json:"refID"`
	IsOwner      bool      `json:"isOwner"`
}

type VaultFolderResponse struct {
	Items         []VaultItem   `json:"items"`
	Folders       []VaultFolder `json:"folders"`
	CurrentFolder VaultFolder   `json:"folder"`
	KeySequence   [][]byte      `json:"keySequence"`
}

type VaultDownloadResponse struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Size         int    `json:"size"`
	Chunks       int    `json:"chunks"`
	ProtectedKey []byte `json:"protectedKey"`
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
	Identifier             string `json:"identifier"`
	LoginKeyHash           string `json:"loginKeyHash"`
	PublicKey              []byte `json:"publicKey"`
	ProtectedKey           []byte `json:"protectedKey"`
	ProtectedRootFolderKey []byte `json:"protectedRootFolderKey"`
}

type SignupResponse struct {
	Identifier string `json:"identifier"`
	Captcha    string `json:"captcha"`
	Error      string `json:"error"`
}

type VerifyAccount struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	LoginKeyHash  []byte `json:"loginKeyHash"`
	ProtectedKey  []byte `json:"protectedKey"`
	PublicKey     []byte `json:"publicKey"`
	RootFolderKey []byte `json:"rootFolderKey"`
}

type Login struct {
	Identifier   string `json:"identifier"`
	LoginKeyHash []byte `json:"loginKeyHash"`
}

type LoginResponse struct {
	PublicKey    []byte `json:"publicKey"`
	ProtectedKey []byte `json:"protectedKey"`
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

type PubKeyResponse struct {
	PublicKey []byte `json:"publicKey"`
}

type ShareItemRequest struct {
	User         string `json:"user"`
	CanModify    bool   `json:"canModify"`
	ProtectedKey []byte `json:"protectedKey"`
}

type NewSharedItem struct {
	ItemID       string
	UserID       string
	SharerName   string
	RecipientID  string
	ProtectedKey []byte
	CanModify    bool
}

type FileOwnershipInfo struct {
	CanModify bool `json:"canModify"`
}

type FolderOwnershipInfo struct {
	ID        string `json:"id"`
	RefID     string `json:"refID"`
	CanModify bool   `json:"canModify"`
	IsOwner   bool   `json:"isOwner"`
}

type ShareInfo struct {
	ID        string `json:"id"`
	Recipient string `json:"recipientName"`
	CanModify bool   `json:"canModify"`
}

type ShareEdit struct {
	ID        string `json:"id"`
	ItemID    string `json:"itemID"`
	CanModify bool   `json:"canModify"`
}

type DeleteResponse struct {
	FreedSpace int `json:"freedSpace"`
}

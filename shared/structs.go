package shared

import "time"

type AccountResponse struct {
	Email              string    `json:"email"`
	StorageAvailable   int       `json:"storageAvailable"`
	StorageUsed        int       `json:"storageUsed"`
	SendAvailable      int       `json:"sendAvailable"`
	SendUsed           int       `json:"sendUsed"`
	SubscriptionExp    time.Time `json:"subscriptionExp" ts_type:"Date" ts_transform:"new Date(__VALUE__)"`
	SubscriptionMethod string    `json:"subscriptionMethod"`
}

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

type ModifyVaultItem struct {
	Name string `json:"name"`
}

type MetadataUploadResponse struct {
	ID string `json:"id"`
}

type NewFolderResponse struct {
	ID string `json:"id"`
}

type VaultItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Size         int       `json:"size"`
	Modified     time.Time `json:"modified" ts_type:"Date" ts_transform:"new Date(__VALUE__)"`
	ProtectedKey []byte    `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	SharedWith   int       `json:"sharedWith"`
	SharedBy     string    `json:"sharedBy"`
	LinkTag      string    `json:"linkTag"`
	CanModify    bool      `json:"canModify"`
	IsOwner      bool      `json:"isOwner"`
	RefID        string    `json:"refID"`
}

type NewVaultFolder struct {
	Name         string `json:"name"`
	ProtectedKey []byte `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	ParentID     string `json:"parentID"`
}

type NewPublicVaultFolder struct {
	ID           string `json:"id"`
	ProtectedKey []byte `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	LinkTag      string `json:"linkTag"`
}

type VaultFolder struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Modified     time.Time `json:"modified" ts_type:"Date" ts_transform:"new Date(__VALUE__)"`
	ParentID     string    `json:"parentID"`
	ProtectedKey []byte    `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
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
	KeySequence   [][]byte      `json:"keySequence" ts_type:"Uint8Array[]" ts_transform:"__VALUE__.map(base64ToArray)"`
}

type VaultDownloadResponse struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Size         int    `json:"size"`
	Chunks       int    `json:"chunks"`
	ProtectedKey []byte `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
}

type PlaintextUpload struct {
	Name       string `json:"name"`
	Salt       []byte `json:"salt" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	Downloads  int    `json:"downloads"`
	Expiration string `json:"expiration"`
	Text       []byte `json:"text" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
}

type DownloadResponse struct {
	Name       string    `json:"name"`
	ID         string    `json:"id"`
	Salt       []byte    `json:"salt" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	Size       int       `json:"size"`
	Chunks     int       `json:"chunks"`
	Downloads  int       `json:"downloads"`
	Expiration time.Time `json:"expiration" ts_type:"Date" ts_transform:"new Date(__VALUE__)"`
}

type Signup struct {
	Identifier     string `json:"identifier"`
	LoginKeyHash   []byte `json:"loginKeyHash" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	PublicKey      []byte `json:"publicKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	ProtectedKey   []byte `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	RootFolderKey  []byte `json:"rootFolderKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	ServerPassword string `json:"serverPassword"`
}

type SignupResponse struct {
	Identifier string `json:"identifier"`
	Captcha    string `json:"captcha"`
	Error      string `json:"error"`
}

type VerifyAccount struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	LoginKeyHash  []byte `json:"loginKeyHash" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	ProtectedKey  []byte `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	PublicKey     []byte `json:"publicKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	RootFolderKey []byte `json:"rootFolderKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
}

type Login struct {
	Identifier   string `json:"identifier"`
	LoginKeyHash []byte `json:"loginKeyHash" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
}

type LoginResponse struct {
	PublicKey    []byte `json:"publicKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
	ProtectedKey []byte `json:"protectedKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
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
	ConfirmPassword string `json:"confirmPassword"`
}

type PubKeyResponse struct {
	PublicKey []byte `json:"publicKey" ts_type:"Uint8Array" ts_transform:"__VALUE__ ? base64ToArray(__VALUE__) : new Uint8Array()"`
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

type DeleteAccount struct {
	Identifier string `json:"identifier"`
}

package models

import "time"

type VaultItem struct {
	ID           string
	RefID        string
	Name         string
	IsFolder     bool
	Size         int
	Modified     time.Time
	SharedWith   int
	SharedBy     string
	IsOwner      bool
	CanModify    bool
	ProtectedKey []byte
}

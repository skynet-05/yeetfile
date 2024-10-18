package models

import (
	"time"
	"yeetfile/shared"
)

type VaultItem struct {
	ID           string
	RefID        string
	Name         string
	IsFolder     bool
	Size         int64
	Modified     time.Time
	SharedWith   int
	SharedBy     string
	IsOwner      bool
	CanModify    bool
	ProtectedKey []byte
	PassEntry    shared.PassEntry
}

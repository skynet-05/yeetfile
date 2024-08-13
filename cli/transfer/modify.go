package transfer

import (
	"yeetfile/cli/globals"
	"yeetfile/shared"
)

// DeleteItem deletes either a file or folder from the user's vault
func DeleteItem(itemID string, isShared, isFolder bool) error {
	if isFolder {
		return globals.API.DeleteVaultFolder(itemID, isShared)
	} else {
		return globals.API.DeleteVaultFile(itemID, isShared)
	}
}

// RenameItem renames a file or folder to the user's specified new name. This
// name is encrypted and then encoded in hex before this function is called.
func RenameItem(itemID, hexEncName string, isFolder bool) error {
	mod := shared.ModifyVaultItem{Name: hexEncName}
	if isFolder {
		return globals.API.ModifyVaultFolder(itemID, mod)
	} else {
		return globals.API.ModifyVaultFile(itemID, mod)
	}
}

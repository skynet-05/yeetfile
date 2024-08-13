package share

import (
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/shared"
)

const ReadPerm = "Read Only"
const WritePerm = "Read + Write"

type Action int

const (
	Cancel Action = iota
	Edit
	Remove
	Add
)

type Perm int

const (
	Read Perm = iota
	Write
)

func fetchSharedInfo(item models.VaultItem) ([]shared.ShareInfo, error) {
	if item.IsFolder {
		return globals.API.GetSharedFolderInfo(item.ID)
	} else {
		return globals.API.GetSharedFileInfo(item.ID)
	}
}

func removeAccess(
	item models.VaultItem,
	shares []shared.ShareInfo,
) ([]shared.ShareInfo, error) {
	if item.IsFolder {
		return globals.API.RemoveSharedFolderUsers(item.ID, shares)
	} else {
		return globals.API.RemoveSharedFileUsers(item.ID, shares)
	}
}

func editPermissions(
	item models.VaultItem,
	shares []shared.ShareInfo,
) ([]shared.ShareInfo, error) {
	if item.IsFolder {
		return globals.API.UpdateSharedFolderUsers(item.ID, shares)
	} else {
		return globals.API.UpdateSharedFileUsers(item.ID, shares)
	}
}

func shareItem(
	item models.VaultItem,
	decryptFunc crypto.CryptFunc,
	decryptKey []byte,
	recipient string,
	perm Perm,
) (shared.ShareInfo, error) {
	itemKey, err := decryptFunc(decryptKey, item.ProtectedKey)
	if err != nil {
		return shared.ShareInfo{}, err
	}

	userKey, err := generateUserProtectedKey(recipient, itemKey)
	if err != nil {
		return shared.ShareInfo{}, err
	}

	shareRequest := shared.ShareItemRequest{
		User:         recipient,
		CanModify:    perm == Write,
		ProtectedKey: userKey,
	}

	if item.IsFolder {
		return globals.API.ShareFolderWithUser(shareRequest, item.ID)
	} else {
		return globals.API.ShareFileWithUser(shareRequest, item.ID)
	}
}

func generateUserProtectedKey(
	recipient string,
	key []byte,
) ([]byte, error) {
	pubKeyResponse, err := globals.API.FetchUserPubKey(recipient)
	if err != nil {
		return nil, err
	}

	userItemKey, err := crypto.EncryptRSA(pubKeyResponse.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return userItemKey, nil
}

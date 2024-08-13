//go:build server_test

package api

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
	"yeetfile/cli/crypto"
	"yeetfile/shared"
)

func prepSharedContent(
	user TestUser,
	key []byte,
	canModify bool,
	recipient string,
) (shared.ShareItemRequest, error) {
	resp, err := user.context.FetchUserPubKey(recipient)
	if err != nil {
		return shared.ShareItemRequest{}, err
	}

	protectedKey, err := crypto.EncryptRSA(resp.PublicKey, key)
	if err != nil {
		return shared.ShareItemRequest{}, err
	}

	return shared.ShareItemRequest{
		User:         recipient,
		ProtectedKey: protectedKey,
		CanModify:    canModify,
	}, nil
}

func TestShareFile(t *testing.T) {
	id, _ := uploadRandomFile(UserA, "", nil)
	_, err := UserB.context.GetVaultItemMetadata(id)
	assert.NotNil(t, err)

	meta, _ := UserA.context.GetVaultItemMetadata(id)
	key, _ := crypto.DecryptRSA(UserA.privKey, meta.ProtectedKey)

	request, err := prepSharedContent(UserA, key, false, UserB.id)
	assert.Nil(t, err)

	share, err := UserA.context.ShareFileWithUser(request, id)
	assert.Nil(t, err)

	_, err = UserB.context.GetVaultItemMetadata(id)
	assert.Nil(t, err)

	err = UserB.context.ModifyVaultFile(id, shared.ModifyVaultItem{Name: "_"})
	assert.NotNil(t, err)

	share.CanModify = true
	_, err = UserA.context.UpdateSharedFileUsers(id, []shared.ShareInfo{share})
	assert.Nil(t, err)

	newName := []byte("new file name")
	encName, _ := crypto.EncryptChunk(key, newName)
	hexName := hex.EncodeToString(encName)
	err = UserB.context.ModifyVaultFile(id, shared.ModifyVaultItem{Name: hexName})
	assert.Nil(t, err)

	meta, _ = UserA.context.GetVaultItemMetadata(id)
	nameBytes, _ := hex.DecodeString(meta.Name)
	decName, _ := crypto.DecryptChunk(key, nameBytes)

	assert.Equal(t, newName, decName)
}

func TestShareFolder(t *testing.T) {
	folderKey, folderID, err := createRandomFolder(UserA, "", nil)
	assert.Nil(t, err)

	fileID, err := uploadRandomFile(UserA, folderID, folderKey)
	assert.Nil(t, err)

	_, err = UserB.context.FetchFolderContents(fileID)
	assert.NotNil(t, err)

	shareRequest, err := prepSharedContent(UserA, folderKey, false, UserB.id)
	assert.Nil(t, err)

	shareInfo, err := UserA.context.ShareFolderWithUser(shareRequest, folderID)
	assert.Nil(t, err)

	folderContents, err := UserB.context.FetchFolderContents(folderID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(folderContents.Items))

	file := folderContents.Items[0]
	userBKeyPair := crypto.IngestKeys(UserB.privKey, UserB.pubKey)
	decFolderKey, err := userBKeyPair.UnwindKeySequence(folderContents.KeySequence)
	assert.Nil(t, err)

	fileKey, err := crypto.DecryptChunk(decFolderKey, file.ProtectedKey)
	assert.Nil(t, err)

	encFileName, _ := hex.DecodeString(file.Name)
	fileName, err := crypto.DecryptChunk(fileKey, encFileName)
	assert.Nil(t, err)
	assert.True(t, len(fileName) > 0)

	_, err = uploadRandomFile(UserB, folderID, folderKey)
	assert.NotNil(t, err) // CanModify is set to false, should reject upload

	shareInfo.CanModify = true
	_, err = UserA.context.UpdateSharedFolderUsers(folderID, []shared.ShareInfo{shareInfo})
	assert.Nil(t, err)

	newFileID, err := uploadRandomFile(UserB, folderID, folderKey)
	assert.Nil(t, err)

	userAContents, err := UserA.context.FetchFolderContents(folderID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(userAContents.Items))

	hasFileID := false
	for _, item := range userAContents.Items {
		hasFileID = hasFileID || newFileID == item.ID
	}

	assert.True(t, hasFileID)

	err = UserB.context.DeleteVaultFolder(folderID, false)
	assert.NotNil(t, err)

	newName := []byte("New Folder Name")
	encName, _ := crypto.EncryptChunk(folderKey, newName)
	hexEncName := hex.EncodeToString(encName)
	err = UserB.context.ModifyVaultFolder(folderID, shared.ModifyVaultItem{Name: hexEncName})
	assert.Nil(t, err)

	folderInfo, err := UserA.context.FetchFolderContents(folderID)
	assert.Nil(t, err)
	encNameBytes, _ := hex.DecodeString(folderInfo.CurrentFolder.Name)
	decNameBytes, _ := crypto.DecryptChunk(folderKey, encNameBytes)
	assert.Equal(t, newName, decNameBytes)

	_, err = UserA.context.RemoveSharedFolderUsers(folderID, []shared.ShareInfo{shareInfo})
	assert.Nil(t, err)

	_, err = UserB.context.FetchFolderContents(folderID)
	assert.NotNil(t, err)
}

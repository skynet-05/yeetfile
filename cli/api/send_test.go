//go:build server_test

package api

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

func TestSendPastLimit(t *testing.T) {
	account, err := UserA.context.GetAccountInfo()
	assert.Nil(t, err)

	used := account.SendUsed
	fakeSize := account.SendAvailable - account.SendUsed
	realSize := fakeSize + 1

	filename := []byte("too_big")
	contents := []byte(strings.Repeat(".", realSize))
	password := []byte("password")

	key, salt, err := crypto.DeriveSendingKey(password, nil)
	assert.Nil(t, err)

	encData, err := crypto.EncryptChunk(key, contents)
	assert.Nil(t, err)

	encName, _ := crypto.EncryptChunk(key, filename)
	hexName := hex.EncodeToString(encName)

	// Attempt uploading with true size
	uploadMetadata := shared.UploadMetadata{
		Name:       hexName,
		Chunks:     1,
		Size:       realSize,
		Salt:       salt,
		Downloads:  1,
		Expiration: "10m",
	}

	_, err = UserA.context.InitSendFile(uploadMetadata)
	assert.NotNil(t, err)

	// Update size to make the init succeed
	uploadMetadata.Size = fakeSize
	meta, err := UserA.context.InitSendFile(uploadMetadata)
	assert.Nil(t, err)

	// Attempt to upload a chunk larger than user has room for
	uploadURL := endpoints.UploadSendFileData.Format(server, meta.ID, "1")
	_, err = UserA.context.UploadFileChunk(uploadURL, encData)
	assert.NotNil(t, err)

	account, err = UserA.context.GetAccountInfo()
	assert.Nil(t, err)
	assert.Equal(t, used, account.SendUsed)
}

func TestSendFile(t *testing.T) {
	filename := []byte("abc123")
	contents := []byte("testing")
	password := []byte("password")

	key, salt, err := crypto.DeriveSendingKey(password, nil)
	if err != nil {
		t.Fatalf("Error deriving sending key: %v\n", key)
	}

	encData, err := crypto.EncryptChunk(key, contents)
	if err != nil {
		t.Fatalf("Error encrypting file data")
	}

	encName, err := crypto.EncryptChunk(key, filename)
	if err != nil {
		t.Fatalf("Error encrypting file chunk")
	}

	hexName := hex.EncodeToString(encName)

	meta, err := UserA.context.InitSendFile(shared.UploadMetadata{
		Name:       hexName,
		Chunks:     1,
		Size:       len(encData),
		Salt:       salt,
		Downloads:  2,
		Expiration: "5s", // 5 seconds
	})
	if err != nil {
		t.Fatalf("Error initializing send file")
	}

	uploadURL := endpoints.UploadSendFileData.Format(server, meta.ID, "1")
	id, err := UserA.context.UploadFileChunk(uploadURL, encData)
	if err != nil {
		t.Fatalf("Error uploading file chunk: %v\n", err)
	} else if meta.ID != id {
		t.Fatalf("Send file metadata ID doesn't match upload ID")
	}

	downloadUrl := endpoints.DownloadSendFileData.Format(server, meta.ID, "1")
	encDownloadedData, err := UserB.context.DownloadFileChunk(downloadUrl)
	if err != nil {
		t.Fatalf("Error downloading send file data: %v\n", err)
	}

	newKey, _, _ := crypto.DeriveSendingKey(password, salt)
	downloadedData, err := crypto.DecryptChunk(newKey, encDownloadedData)
	if err != nil {
		t.Fatalf("Error decrypting downloaded data")
	}

	if string(downloadedData) != string(contents) {
		t.Fatalf("Decrypted content does not match ('%s' vs '%s')",
			string(downloadedData),
			string(contents))
	}

	// Wait for file to expire (+1s buffer to avoid race)
	log.Println("Waiting for file to expire...")
	time.Sleep(6 * time.Second)

	_, err = UserA.context.FetchSendFileMetadata(server, meta.ID)
	if err == nil {
		t.Fatal("User was able to download sent file after expiration")
	}
}

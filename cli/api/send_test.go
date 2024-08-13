//go:build server_test

package api

import (
	"encoding/hex"
	"log"
	"testing"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

func TestSendFile(t *testing.T) {
	filename := []byte("abc123")
	contents := []byte("testing")
	password := []byte("password")

	key, salt, pepper, err := crypto.DeriveSendingKey(password, nil, nil)
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

	newKey, _, _, _ := crypto.DeriveSendingKey(password, salt, pepper)
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

package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var failedKeyGen = errors.New("failed to derive key")
var wrongPassword = errors.New("incorrect password")
var failedDecrypt = errors.New("failed to decrypt data")

type FileDownload struct {
	ID         string
	Name       string
	Size       int
	Chunks     int
	Expiration time.Time
	Downloads  int
	Key        [shared.KeySize]byte
}

// FetchMetadata retrieves file metadata for the requested file path
func FetchMetadata(path string) (shared.DownloadResponse, error) {
	url := fmt.Sprintf("%s/d/%s", userConfig.Server, path)
	resp, err := GetRequest(url)
	if err != nil {
		return shared.DownloadResponse{}, err
	}

	decoder := json.NewDecoder(resp.Body)
	var d shared.DownloadResponse
	err = decoder.Decode(&d)
	if err != nil {
		return shared.DownloadResponse{}, err
	}

	return d, nil
}

// PrepareDownload prepares for downloading a file by ensuring that a valid
// key is available to decrypt file chunks. Returns a FileDownload struct that
// can be used to start downloading the file.
func PrepareDownload(
	d shared.DownloadResponse,
	pw []byte,
	pepper []byte,
) (FileDownload, error) {
	key, _, _, err := crypto.DeriveKey(pw, d.Salt, pepper)
	if err != nil {
		return FileDownload{}, failedKeyGen
	}

	// Attempt to decrypt the filename in order to check the key's validity
	name, _ := hex.DecodeString(d.Name)
	decName, err := crypto.DecryptString(key, name)

	if err != nil {
		return FileDownload{}, wrongPassword
	}

	return FileDownload{
		ID:     d.ID,
		Name:   decName,
		Chunks: d.Chunks,
	}, nil
}

// VerifyDownload displays file metadata to the user to ensure that the file
// is what they're expecting.
func (file FileDownload) VerifyDownload() bool {
	timeDiff := time.Now().Sub(file.Expiration)

	fmt.Println(utils.LineDecorator)
	fmt.Printf("Name: %s\n", file.Name)
	fmt.Printf("Size: %s", utils.ReadableFileSize(file.Size))
	fmt.Printf("Downloads Remaining: %d\n", file.Downloads)
	fmt.Printf("Expires: %s (%s)\n", file.Expiration, timeDiff)
	fmt.Println(utils.LineDecorator)

	shouldDownload := utils.StringPrompt("Download? (y/n)")
	return strings.ToLower(shouldDownload) == "y"
}

// DownloadFile downloads file contents and decrypts them before saving the file
func (file FileDownload) DownloadFile() error {
	var output []byte

	out, _ := os.OpenFile(file.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)

	i := 0
	for i < file.Chunks {
		fmt.Printf("\033[2K\rDownloading...(%d/%d)", i+1, file.Chunks)

		url := fmt.Sprintf("%s/d/%s/%d", userConfig.Server, file.ID, i+1)
		resp, err := GetRequest(url)
		body, _ := io.ReadAll(resp.Body)

		plaintext, err := crypto.DecryptChunk(file.Key, body)
		if err != nil {
			return failedDecrypt
		}

		output = append(output, plaintext...)
		i += 1
	}

	fmt.Print("\u001B[2K\nDownload finished!\n")

	_, _ = out.Write(output)
	_ = out.Close()

	fmt.Printf("\nOutput: %s\n", file.Name)

	if file.Downloads == 1 {
		// This download was the last one, and the file has been deleted
		fmt.Println("The file has been deleted from the server")
	}

	return nil
}

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"yeetfile/cli/crypto"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

// UploadFile is the entrypoint to uploading a file to the server. It receives
// the filename, number of downloads, and expiration date for a file.
func UploadFile(path string, downloads int, exp string) bool {
	if !hasValidSession() {
		fmt.Println("Login required")

		// Try logging user in and then repeating the request
		if LoginUser() {
			return UploadFile(path, downloads, exp)
		} else {
			fmt.Println("You need to log in before uploading")
			return false
		}
	}

	filename := filepath.Base(path)

	fmt.Println("Uploading file:", filename)
	fmt.Println("==========")

	pw := utils.RequestPassword()
	if !utils.ConfirmPassword(pw) {
		fmt.Println("Passwords do not match")
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		panic("Unable to open file")
	}

	stat, err := file.Stat()

	key, salt, pepper, err := crypto.DeriveKey(pw, nil, nil)

	// Encrypt and encode the file name (encoding required for upload to B3)
	encName := crypto.EncryptChunk(key, []byte(filename))
	hexEncName := hex.EncodeToString(encName)

	id, err := InitializeUpload(hexEncName, salt, stat.Size(), downloads, exp)

	if len(id) > 0 {
		var path string
		if stat.Size() > int64(shared.ChunkSize) {
			path, err = MultiPartUpload(id, file, stat.Size(), key)
		} else {
			path, err = SingleUpload(id, file, stat.Size(), key)
		}

		if err != nil {
			fmt.Printf("Error uploading file: %v\n", err)
			return false
		} else {
			fmt.Printf("\nResource: %s#%s\n", path, string(pepper))
			fmt.Printf("Link: %s/%s#%s\n", userConfig.Server, path, string(pepper))
			return true
		}
	} else {
		fmt.Println("Server returned an invalid file id")
		return false
	}
}

// InitializeUpload begins the upload process by sending the server metadata
// about the file. This includes the name of the file (encrypted and hex
// encoded), the salt, the length, the number of downloads allowed, and the
// date that the file should expire.
func InitializeUpload(
	hexEncName string,
	salt []byte,
	length int64,
	downloads int,
	exp string,
) (string, error) {
	fmt.Print("\033[2K\rInitializing upload...")

	numChunks := math.Ceil(float64(length) / float64(shared.ChunkSize))

	uploadMetadata := shared.UploadMetadata{
		Name:       hexEncName,
		Salt:       salt,
		Chunks:     int(numChunks),
		Downloads:  downloads,
		Expiration: exp,
	}

	reqData, err := json.Marshal(uploadMetadata)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/u", userConfig.Server)
	resp, err := PostRequest(url, reqData)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("\033[2K\r\nServer response: %d\n", resp.StatusCode)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body: ", err)
		return "", err
	}

	fmt.Print("\033[2K\rInitializing upload: DONE")
	fmt.Println()

	return string(body), nil
}

// MultiPartUpload uploads a file in multiple chunks, with each chunk containing
// at most the value of shared.ChunkSize (5mb). The function requires an ID from
// InitializeUpload, the file pointer, the file size, and the key for encryption
func MultiPartUpload(id string, file *os.File, size int64, key [32]byte) (string, error) {
	fmt.Print("\033[2K\rUploading...")

	var path string
	i := 0
	start := int64(0)
	for start < size {
		start = int64(shared.ChunkSize * i)
		end := int64(shared.ChunkSize * (i + 1))

		if end > size {
			end = size
		}

		contents := make([]byte, end-start)
		_, err := file.ReadAt(contents, start)
		encData := crypto.EncryptChunk(key, contents)

		url := fmt.Sprintf("%s/u/%s/%d", userConfig.Server, id, i+1)
		resp, err := PostRequest(url, encData)
		if err != nil {
			fmt.Println("Error sending data")
			return "", err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error fetching response")
			return "", err
		}

		if len(body) > 0 {
			fmt.Print("\033[2K\rUploading: DONE")
			fmt.Println()
			path = string(body)
			break
		}

		i += 1
	}

	return path, nil
}

// SingleUpload uploads a file's contents in one chunk. This can only be done if
// the total file size is less than the chunk size (5mb). The function requires
// an ID, the file pointer, the file length, and the key for encryption.
func SingleUpload(id string, file *os.File, length int64, key [32]byte) (string, error) {
	fmt.Print("\033[2K\rUploading...")

	content := make([]byte, length)
	size, err := file.Read(content)
	if err != nil || int64(size) != length {
		fmt.Println("Error reading file")
		return "", err
	}

	data := crypto.EncryptChunk(key, content)

	url := fmt.Sprintf("%s/u/%s/1", userConfig.Server, id)
	resp, err := PostRequest(url, data)
	if err != nil {
		fmt.Println("Error sending data")
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error fetching response")
		return "", err
	}

	fmt.Print("\033[2K\rUploading: DONE")
	fmt.Println()
	return string(body), nil
}

func hasValidSession() bool {
	url := fmt.Sprintf("%s/session", userConfig.Server)
	resp, err := GetRequest(url)
	if err != nil {
		return false
	}

	return resp.StatusCode == http.StatusOK

}

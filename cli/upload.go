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

// StartFileUpload is the entrypoint to uploading a file to the server. It
// receives the filename, number of downloads, and expiration date for a file.
func StartFileUpload(path string, downloads int, exp string) bool {
	filename := filepath.Base(path)

	fmt.Println("Uploading file:", filename)
	fmt.Println("==========")

	file, err := os.Open(path)
	if err != nil {
		panic("Unable to open file")
	}

	stat, err := file.Stat()

	key, salt, pepper, err := generateKey()

	// Encrypt and encode the file name (encoding required for upload to B3)
	encName := crypto.EncryptChunk(key, []byte(filename))
	hexEncName := hex.EncodeToString(encName)

	numChunks := math.Ceil(float64(stat.Size()) / float64(shared.ChunkSize))

	metadata := shared.UploadMetadata{
		Name:       hexEncName,
		Salt:       salt,
		Size:       int(stat.Size()),
		Chunks:     int(numChunks),
		Downloads:  downloads,
		Expiration: exp,
	}

	id, err := uploadMetadata(metadata)

	if len(id) > 0 {
		var path string
		if stat.Size() > int64(shared.ChunkSize) {
			path, err = UploadMultiChunk(id, file, stat.Size(), key)
		} else {
			content := make([]byte, stat.Size())
			_, err = file.Read(content)
			path, err = UploadSingleChunk(id, content, key)
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

// StartPlaintextUpload uploads ASCII text to the server. This is distinct from
// uploading a file, since it doesn't affect the user's transfer limit, but is
// limited to shared.MaxPlaintextLen characters.
func StartPlaintextUpload(text string, downloads int, exp string) bool {
	if len(text) > shared.MaxPlaintextLen {
		fmt.Printf("Error: Text exceeds %d characters and should "+
			"be uploaded as a file", shared.MaxPlaintextLen)
		return false
	} else if !shared.IsPlaintext(text) {
		fmt.Println("Error: Text contains non-ASCII characters and should " +
			"be uploaded as a file")
		return false
	}

	key, salt, pepper, err := generateKey()
	encName := crypto.EncryptChunk(key, []byte(shared.GenRandomString(10)))
	hexEncName := hex.EncodeToString(encName)
	plaintextUpload := shared.PlaintextUpload{
		Name:       hexEncName,
		Salt:       salt,
		Downloads:  downloads,
		Expiration: exp,
	}

	path, err := UploadPlaintext([]byte(text), key, plaintextUpload)

	if err != nil {
		fmt.Printf("Error uploading text: %v\n", err)
		return false
	} else {
		fmt.Printf("\nResource: %s#%s\n", path, string(pepper))
		fmt.Printf("Link: %s/%s#%s\n", userConfig.Server, path, string(pepper))
		return true
	}
}

// generateKey prompts the user for a password (blank pw is ok) and uses that to
// derive a key for the content that is being uploaded
func generateKey() ([shared.KeySize]byte, []byte, []byte, error) {
	pw := utils.RequestPassword()
	if !utils.ConfirmPassword(pw) {
		fmt.Printf("\nError: passwords do not match!\n\n")
		return generateKey()
	}

	return crypto.DeriveKey(pw, nil, nil)
}

// uploadMetadata begins the upload process by sending the server metadata
// about the file. This includes the name of the file (encrypted and hex
// encoded), the salt, the length, the number of downloads allowed, and the
// date that the file should expire.
func uploadMetadata(meta shared.UploadMetadata) (string, error) {
	fmt.Print("\033[2K\rInitializing upload...")

	reqData, err := json.Marshal(meta)
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

// UploadMultiChunk uploads a file in multiple chunks, with each chunk containing
// at most the value of shared.ChunkSize (5mb). The function requires an ID from
// InitializeUpload, the file pointer, the file size, and the key for encryption
func UploadMultiChunk(id string, file *os.File, size int64, key [32]byte) (string, error) {
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

// UploadSingleChunk uploads a file's contents in one chunk. This can only be
// done if the total file size is less than the chunk size (5mb). The function
// requires an ID, the file content, and the key for encryption.
func UploadSingleChunk(id string, content []byte, key [shared.KeySize]byte) (string, error) {
	fmt.Print("\033[2K\rUploading...")

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

// UploadPlaintext uploads ASCII text to the server in a single chunk. The
// endpoint for this request doesn't require authentication, but is limited to
// shared.MaxPlaintextLen characters (shared/constants.go).
func UploadPlaintext(
	content []byte,
	key [shared.KeySize]byte,
	info shared.PlaintextUpload,
) (string, error) {
	info.Text = crypto.EncryptChunk(key, content)

	reqData, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/plaintext", userConfig.Server)
	resp, err := PostRequest(url, reqData)

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

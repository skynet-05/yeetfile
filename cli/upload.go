package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"yeetfile/crypto"
	"yeetfile/shared"
)

// UploadFile is the entrypoint to uploading a file to the server. It receives
// the filename, number of downloads, and expiration date for a file.
func UploadFile(filename string, downloads int, exp string) {
	fmt.Println("Uploading file:", filename)
	fmt.Println("==========")

	fmt.Print("Enter Password: ")
	pw, err := term.ReadPassword(syscall.Stdin)

	fmt.Print("\nConfirm Password: ")
	confirm, err := term.ReadPassword(syscall.Stdin)
	fmt.Print("\n")

	if err != nil {
		fmt.Println("Error reading stdin")
		return
	} else if string(pw) != string(confirm) {
		fmt.Println("Passwords don't match")
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		panic("Unable to open file")
	}

	stat, err := file.Stat()

	key, salt, err := crypto.DeriveKey(pw, nil)
	encName := crypto.EncryptChunk(key, []byte(filename))
	hexEncName := hex.EncodeToString(encName)

	id, err := InitializeUpload(hexEncName, salt, stat.Size(), downloads, exp)

	if len(id) > 0 {
		if stat.Size() > int64(shared.ChunkSize) {
			MultiPartUpload(id, file, stat.Size(), key)
		} else {
			SingleUpload(id, file, stat.Size(), key)
		}
	}
}

// InitializeUpload begins the upload process by sending the server metadata
// about the file. This includes the name of the file (encrypted and hex
// encoded), the salt, the length, the number of downloads allowed, and the date
// that the file should expire.
func InitializeUpload(
	hexEncName string,
	salt []byte,
	length int64,
	downloads int,
	exp string,
) (string, error) {
	fmt.Print("\033[2K\rInitializing upload...")
	client := &http.Client{}

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

	url := fmt.Sprintf("%s/u", domain)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
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
func MultiPartUpload(id string, file *os.File, size int64, key [32]byte) {
	client := &http.Client{}

	fmt.Print("\033[2K\rUploading...")

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
		buf := bytes.NewBuffer(encData)
		req, _ := http.NewRequest("POST", domain+"/u/"+id, buf)

		req.Header = http.Header{
			"Chunk": {strconv.Itoa(i + 1)},
		}

		resp, _ := client.Do(req)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error fetching response")
			return
		}

		if len(body) > 0 {
			fmt.Print("\033[2K\rUploading: DONE")
			fmt.Println()
			fmt.Println("Link: " + string(body))
			break
		}

		i += 1
	}
}

// SingleUpload uploads a file's contents in one chunk. This can only be done if
// the total file size is less than the chunk size (5mb). The function requires
// an ID, the file pointer, the file length, and the key for encryption.
func SingleUpload(id string, file *os.File, length int64, key [32]byte) {
	client := &http.Client{}

	fmt.Print("\033[2K\rUploading...")

	content := make([]byte, length)
	size, err := file.Read(content)
	if err != nil || int64(size) != length {
		fmt.Println("Error reading file")
		return
	}

	data := crypto.EncryptChunk(key, content)
	buf := bytes.NewBuffer(data)
	req, _ := http.NewRequest("POST", domain+"/u/"+id, buf)

	req.Header = http.Header{
		"Chunk": {"1"},
	}

	resp, _ := client.Do(req)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error fetching response")
		return
	}

	fmt.Print("\033[2K\rUploading: DONE")
	fmt.Println()
	fmt.Println("Link: " + string(body))
}

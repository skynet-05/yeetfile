package main

import (
	"bytes"
	"fmt"
	"golang.org/x/term"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"syscall"
)

type Upload struct {
	ID   string
	Name string
	Key  string
	Data []byte
}

func UploadFile(filename string) {
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

	file, err := os.ReadFile(filename)
	if err != nil {
		panic("Unable to open file")
	}

	upload := InitializeUpload(filename, file, string(pw))
	fmt.Println("ID:", upload.ID)

	if len(file) > ChunkSize {
		upload.MultiPartUpload()
	} else {
		upload.SingleUpload()
	}
}

func InitializeUpload(
	filename string,
	data []byte,
	password string,
) Upload {
	fmt.Print("\033[2K\rInitializing upload...")
	client := &http.Client{}

	numChunks := math.Ceil(float64(len(data)) / float64(ChunkSize))

	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"name": "%s",
		"chunks": %d,
		"password": "%s"
	}`, filename, int(numChunks), password)))

	req, err := http.NewRequest("POST", domain+"/u", reqBody)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return Upload{}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return Upload{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		return Upload{}
	}

	fmt.Print("\033[2K\rInitializing upload: DONE")
	fmt.Println()

	// TODO: Update server to return json instead of string response
	response := strings.Split(string(body), "|")

	return Upload{
		ID:   response[0],
		Key:  response[1],
		Name: filename,
		Data: data,
	}
}

func (upload Upload) MultiPartUpload() {

}

func (upload Upload) SingleUpload() {

}

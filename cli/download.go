package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/term"
	"io"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
	"yeetfile/crypto"
	"yeetfile/shared"
)

var wrongPassword = errors.New("incorrect password")
var failedDecrypt = errors.New("failed to decrypt data")

// StartDownload initiates a download of a file using the file's human-readable
// tag, a 3-word period-separated string such as "machine.delirium.yarn" (which
// is returned when uploading a file).
func StartDownload(tag string) {
	client := &http.Client{}

	fmt.Print("Enter Password: ")
	pw, _ := term.ReadPassword(syscall.Stdin)
	fmt.Println()

	req, _ := http.NewRequest("GET", domain+"/d/"+tag, nil)

	resp, _ := client.Do(req)
	decoder := json.NewDecoder(resp.Body)
	var d shared.DownloadResponse
	_ = decoder.Decode(&d)

	key, _, err := crypto.DeriveKey(pw, d.Salt)
	if err != nil {
		fmt.Println("Failed to derive key")
		return
	}

	err = DownloadFile(d, key)
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
}

// DownloadFile downloads file contents and decrypts them before saving the file
func DownloadFile(d shared.DownloadResponse, key [32]byte) error {
	client := &http.Client{}
	var output []byte

	name, _ := hex.DecodeString(d.Name)
	decName, err := crypto.DecryptString(key, name)

	if err != nil {
		return wrongPassword
	}

	out, _ := os.OpenFile(decName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)

	i := 0
	var resp *http.Response
	for i < d.Chunks {
		fmt.Printf("\033[2K\rDownloading...(%d/%d)", i+1, d.Chunks)
		url := fmt.Sprintf("%s/d/%s/%d", domain, d.ID, i+1)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header = http.Header{
			"Chunk": {strconv.Itoa(i + 1)},
		}

		resp, _ = client.Do(req)
		body, _ := io.ReadAll(resp.Body)

		plaintext, _, err := crypto.DecryptChunk(key, body)
		if err != nil {
			return failedDecrypt
		}

		output = append(output, plaintext...)
		i += 1
	}

	fmt.Print("\u001B[2K\nDownload finished!\n")

	_, _ = out.Write(output)
	_ = out.Close()

	showDownloadInfo(resp.Header)

	fmt.Printf("\nOutput: %s\n", decName)
	return nil
}

// showDownloadInfo displays relevant info pertaining to the download to the
// user. This includes the number of downloads remaining and the expiration date
// of the file. This information is encapsulated by the "Downloads" and "Date"
// headers in the download response.
func showDownloadInfo(header http.Header) {
	downloads := header.Get("Downloads")
	date := header.Get("Date")
	remaining := -1

	if len(downloads) > 0 {
		remaining, _ = strconv.Atoi(downloads)
		fmt.Printf("-- Downloads remaining: %d\n", remaining)
		if remaining == 0 {
			fmt.Println("   File has been deleted!")
		}
	}

	if len(date) > 0 && remaining != 0 {
		exampleDate := "2006-01-02 15:04:05.999999999 -0700 MST"
		parse, err := time.Parse(exampleDate, date)
		diff := time.Now().Sub(parse)

		if err != nil {
			fmt.Printf("Error parsing exp date: %v\n", err)
			return
		}
		
		fmt.Printf("-- Expires: %s\n   (%s)\n", date, diff)
	}
}

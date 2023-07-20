package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"io"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"yeetfile/shared"
)

func StartDownload(tag string) {
	client := &http.Client{}

	fmt.Print("Enter Password: ")
	pw, _ := term.ReadPassword(syscall.Stdin)
	fmt.Println()

	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"password": "%s"
	}`, pw)))

	req, _ := http.NewRequest("POST", domain+"/d/"+tag, reqBody)

	resp, _ := client.Do(req)
	decoder := json.NewDecoder(resp.Body)
	var d shared.DownloadResponse
	_ = decoder.Decode(&d)

	DownloadFile(d)
}

func DownloadFile(d shared.DownloadResponse) {
	fmt.Print("\033[2K\rDownloading...")
	client := &http.Client{}
	var output []byte

	out, _ := os.OpenFile(d.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)

	i := 0
	for i < d.Chunks {
		fmt.Printf("\033[2K\rDownloading...(%d/%d)", i+1, d.Chunks)
		url := fmt.Sprintf("%s/d/%s/%d", domain, d.ID, i+1)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header = http.Header{
			"Chunk": {strconv.Itoa(i + 1)},
			"Key":   {d.Key},
		}

		resp, _ := client.Do(req)
		body, _ := io.ReadAll(resp.Body)
		output = append(output, body...)
		i += 1
	}

	fmt.Print("\u001B[2K\nDownload finished!\n")

	_, _ = out.Write(output)
	_ = out.Close()
	fmt.Printf("Output: %s\n", d.Name)
}

package server

import (
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"os"
	"yeetfile/src/b2"
	"yeetfile/src/crypto"
)

func TestDownload() {
	auth, err := b2.AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))

	if err != nil {
		panic(err)
	}

	id := ""
	length := 10035052
	password := []byte("topsecret")

	salt, err := auth.PartialDownloadById(id, length-crypto.KEY_SIZE, length)
	key, _, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return
	}

	// ---------------
	// TODO: Add password validation step before downloading from B2
	// ---------------

	out, err := os.OpenFile("out.enc", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)

	start := 0
	var output []byte
	for start < length-crypto.KEY_SIZE-1 {
		chunkSize := crypto.NONCE_SIZE + crypto.BUFFER_SIZE + secretbox.Overhead + start - 1
		if start+chunkSize > length-crypto.KEY_SIZE-1 {
			chunkSize = length - crypto.KEY_SIZE - 1
		}

		data, _ := auth.PartialDownloadById(id, start, chunkSize)

		plaintext, readSize := crypto.DecryptChunk(key, data)
		output = append(output, plaintext...)
		start += readSize
	}

	_, _ = out.Write(output)
	_ = out.Close()

	plaintext, _ := os.ReadFile("out.enc")
	fmt.Println(string(plaintext))
}

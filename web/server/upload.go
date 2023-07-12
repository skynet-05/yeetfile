package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"yeetfile/b2"
	"yeetfile/crypto"
)

var B2 b2.Auth

type FileUpload struct {
	filename string
	data     []byte
	key      [32]byte
	salt     []byte
}

func TestUpload() {
	filename := "lipsum-big.txt"
	password := []byte("topsecret")

	file, err := os.ReadFile(filename)
	if err != nil {
		panic("Unable to open file")
	}

	key, salt, err := crypto.DeriveKey(password, nil)
	if err != nil {
		log.Fatalf("Failed to derive key: %v", err.Error())
	}

	upload := FileUpload{
		filename: "lipsum-big.enc",
		data:     file,
		key:      key,
		salt:     salt,
	}

	if crypto.BUFFER_SIZE > len(file) {
		upload.UploadFile(0)
	} else {
		upload.UploadLargeFile()
	}

}

func (upload FileUpload) UploadFile(attempts int) {
	info, err := B2.GetUploadURL()
	if err != nil {
		panic(err)
	}

	encData := crypto.EncryptChunk(upload.key, upload.data)
	encData = append(encData, upload.salt...)

	checksum := fmt.Sprintf("%x", crypto.GenChecksum(encData))

	b2File, err := info.UploadFile(
		upload.filename,
		checksum,
		encData,
	)

	if err != nil {
		if attempts < 5 {
			upload.UploadFile(attempts + 1)
		} else {
			log.Fatalf("Unable to upload file")
		}
	}

	fmt.Printf("File ID: %s\n", b2File.FileID)
	fmt.Printf("File size: %d\n", b2File.ContentLength)
}

func (upload FileUpload) UploadLargeFile() {
	init, err := B2.StartLargeFile(upload.filename)
	if err != nil {
		panic(err)
	}

	info, err := B2.GetUploadPartURL(init)
	if err != nil {
		panic(err)
	}

	var checksums []string

	idx := 0
	chunkNum := 1
	for idx < len(upload.data) {
		chunkSize := crypto.BUFFER_SIZE
		needsSalt := false
		if idx+crypto.BUFFER_SIZE > len(upload.data) {
			chunkSize = len(upload.data) - idx
			needsSalt = true
		}

		chunk := crypto.EncryptChunk(
			upload.key,
			upload.data[idx:idx+chunkSize])
		if needsSalt {
			chunk = append(chunk, upload.salt...)
		}
		checksum := crypto.GenChecksum(chunk)
		checksums = append(checksums, fmt.Sprintf("%x", checksum))

		err := info.UploadFilePart(
			chunkNum,
			fmt.Sprintf("%x", checksum),
			chunk,
		)

		if err != nil {
			panic(err)
		}

		idx += chunkSize
		chunkNum += 1
	}

	checksumStr := "[\"" + strings.Join(checksums, "\",\"") + "\"]"

	largeFile, err := B2.FinishLargeFile(info.FileID, checksumStr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("File ID: %s\n", largeFile.FileID)
	fmt.Printf("File size: %d\n", largeFile.ContentLength)
}

func init() {
	var err error
	B2, err = b2.AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))
	if err != nil {
		panic(err)
	}
}

package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"yeetfile/src/b2"
	"yeetfile/src/utils"
)

type FileUpload struct {
	b2       b2.Auth
	filename string
	data     []byte
	key      [32]byte
	salt     []byte
}

func TestUpload() {
	filename := "lipsum-big.txt"
	password := []byte("topsecret")

	b2Auth, err := b2.AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))
	if err != nil {
		panic(err)
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		panic("Unable to open file")
	}

	key, salt, err := utils.DeriveKey(password, nil)
	if err != nil {
		log.Fatalf("Failed to derive key: %v", err.Error())
	}

	upload := FileUpload{
		b2:       b2Auth,
		filename: "lipsum-big.enc",
		data:     file,
		key:      key,
		salt:     salt,
	}

	if utils.BUFFER_SIZE > len(file) {
		upload.UploadFile(0)
	} else {
		upload.UploadLargeFile()
	}

}

func (upload FileUpload) UploadFile(attempts int) {
	info, err := upload.b2.GetUploadURL()
	if err != nil {
		panic(err)
	}

	encData := utils.EncryptChunk(upload.key, upload.data)
	encData = append(encData, upload.salt...)

	checksum := fmt.Sprintf("%x", utils.GenChecksum(encData))

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
	init, err := upload.b2.StartLargeFile(upload.filename)
	if err != nil {
		panic(err)
	}

	info, err := upload.b2.GetUploadPartURL(init)
	if err != nil {
		panic(err)
	}

	var checksums []string

	idx := 0
	chunkNum := 1
	for idx < len(upload.data) {
		chunkSize := utils.BUFFER_SIZE
		needsSalt := false
		if idx+utils.BUFFER_SIZE > len(upload.data) {
			chunkSize = len(upload.data) - idx
			needsSalt = true
		}

		chunk := utils.EncryptChunk(
			upload.key,
			upload.data[idx:idx+chunkSize])
		if needsSalt {
			chunk = append(chunk, upload.salt...)
		}
		checksum := utils.GenChecksum(chunk)
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

	largeFile, err := upload.b2.FinishLargeFile(info.FileID, checksumStr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("File ID: %s\n", largeFile.FileID)
	fmt.Printf("File size: %d\n", largeFile.ContentLength)
}

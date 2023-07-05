package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"yeetfile/src/backblaze"
	"yeetfile/src/utils"
)

type FileUpload struct {
	auth     backblaze.B2Auth
	filename string
	data     []byte
	key      [32]byte
	salt     []byte
}

func TestUpload() {
	filename := "lipsum.txt"
	password := []byte("topsecret")

	b2Auth, err := backblaze.B2Init(os.Getenv("B2_BUCKET_KEY_ID"), os.Getenv("B2_BUCKET_KEY"))
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
		auth:     b2Auth,
		filename: filename,
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
	info, err := upload.auth.B2GetUploadURL()
	if err != nil {
		panic(err)
	}

	encData := utils.EncryptChunk(upload.key, upload.data)
	encData = append(encData, upload.salt...)

	checksum := fmt.Sprintf("%x", utils.GenChecksum(encData))

	err = upload.auth.B2UploadFile(
		info.UploadURL,
		upload.filename,
		checksum,
		encData,
	)

	if err != nil {
		if attempts < 5 {
			upload.UploadFile(attempts + 1)
		} else {
			panic(err)
		}
	}
}

func (upload FileUpload) UploadLargeFile() {
	init, err := upload.auth.B2StartLargeFile("lipsum.enc")
	if err != nil {
		panic(err)
	}

	info, err := upload.auth.B2GetUploadPartURL(init)
	if err != nil {
		panic(err)
	}

	var checksums []string

	idx := 0
	for idx < len(upload.data) {
		chunkSize := utils.BUFFER_SIZE
		needsSalt := false
		if idx+utils.BUFFER_SIZE > len(upload.data) {
			chunkSize = len(upload.data) - idx
			needsSalt = true
		}

		chunk := utils.EncryptChunk(upload.key, upload.data[idx:idx+chunkSize])
		if needsSalt {
			chunk = append(chunk, upload.salt...)
		}
		checksum := utils.GenChecksum(chunk)
		checksums = append(checksums, fmt.Sprintf("%x", checksum))

		err := upload.auth.B2UploadFilePart(
			info.UploadURL,
			idx+1,
			fmt.Sprintf("%x", checksum),
			chunk,
		)

		if err != nil {
			panic(err)
		}

		idx += chunkSize
	}

	checksumStr := "[\"" + strings.Join(checksums, "\",\"") + "\"]"

	err = upload.auth.B2FinishLargeFile(info.FileID, checksumStr)
	if err != nil {
		panic(err)
	}
}

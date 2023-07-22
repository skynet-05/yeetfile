package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"yeetfile/b2"
	"yeetfile/crypto"
	"yeetfile/db"
	"yeetfile/service"
	"yeetfile/utils"
)

type FileUpload struct {
	filename string
	data     []byte
	key      [32]byte
	salt     []byte
	checksum string
	chunk    int
	chunks   int
}

func PrepareUpload(
	id string,
	key [32]byte,
	chunk int,
	data []byte,
) (FileUpload, db.B2Upload) {
	encData := crypto.EncryptChunk(key, data)

	_, checksum := crypto.GenChecksum(encData)
	db.UpdateChecksums(id, checksum)

	metadata := db.RetrieveMetadata(id)
	b2Values := db.GetB2UploadValues(id)

	upload := FileUpload{
		data:     encData,
		filename: metadata.Name,
		key:      key,
		salt:     metadata.Salt,
		checksum: checksum,
		chunk:    chunk,
		chunks:   metadata.Chunks,
	}

	return upload, b2Values
}

func (upload FileUpload) Upload(b2Values db.B2Upload) (bool, error) {
	var err error

	if upload.chunks > 1 {
		largeFile := b2.FilePartInfo{
			FileID:             b2Values.UploadID,
			AuthorizationToken: b2Values.Token,
			UploadURL:          b2Values.UploadURL,
		}

		err = largeFile.UploadFilePart(
			upload.chunk,
			upload.checksum,
			upload.data)

		if err != nil {
			return false, err
		}

		if upload.chunk == upload.chunks {
			b2ID, length := FinishLargeB2Upload(
				b2Values.UploadID,
				utils.StrArrToStr(b2Values.Checksums))
			db.UpdateB2Metadata(b2Values.MetadataID, b2ID, length)
			return true, nil
		} else {
			return false, nil
		}
	} else {
		file := b2.FileInfo{
			BucketID:           b2Values.UploadID,
			AuthorizationToken: b2Values.Token,
			UploadURL:          b2Values.UploadURL,
		}

		resp, err := file.UploadFile(
			upload.filename,
			upload.checksum,
			upload.data)

		if err != nil {
			return false, err
		}

		db.UpdateB2Metadata(
			b2Values.MetadataID,
			resp.FileID,
			resp.ContentLength)

		return true, nil
	}
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

func InitB2Upload() (b2.FileInfo, error) {
	return service.B2.GetUploadURL()
}

func InitLargeB2Upload(filename string) (b2.FilePartInfo, error) {
	init, err := service.B2.StartLargeFile(filename)
	if err != nil {
		panic(err)
	}

	return service.B2.GetUploadPartURL(init)
}

func FinishLargeB2Upload(b2ID string, checksums string) (string, int) {
	largeFile, err := service.B2.FinishLargeFile(b2ID, checksums)
	if err != nil {
		panic(err)
	}

	return largeFile.FileID, largeFile.ContentLength
}

func (upload FileUpload) UploadFile(attempts int) {
	info, err := service.B2.GetUploadURL()
	if err != nil {
		panic(err)
	}

	encData := crypto.EncryptChunk(upload.key, upload.data)
	encData = append(encData, upload.salt...)

	_, checksum := crypto.GenChecksum(encData)

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
	init, err := service.B2.StartLargeFile(upload.filename)
	if err != nil {
		panic(err)
	}

	info, err := service.B2.GetUploadPartURL(init)
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
		_, checksum := crypto.GenChecksum(chunk)
		checksums = append(checksums, checksum)

		err := info.UploadFilePart(
			chunkNum,
			checksum,
			chunk,
		)

		if err != nil {
			panic(err)
		}

		idx += chunkSize
		chunkNum += 1
	}

	checksumStr := "[\"" + strings.Join(checksums, "\",\"") + "\"]"

	largeFile, err := service.B2.FinishLargeFile(info.FileID, checksumStr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("File ID: %s\n", largeFile.FileID)
	fmt.Printf("File size: %d\n", largeFile.ContentLength)
}

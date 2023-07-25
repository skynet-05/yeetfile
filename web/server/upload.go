package server

import (
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

func InitB2Upload() (b2.FileInfo, error) {
	return service.B2.GetUploadURL()
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

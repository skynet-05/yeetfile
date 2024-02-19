package transfer

import (
	"github.com/benbusby/b2"
	db "yeetfile/web/db"
	"yeetfile/web/service"
	"yeetfile/web/utils"
)

type FileUpload struct {
	filename string
	data     []byte
	salt     []byte
	checksum string
	chunk    int
	chunks   int
}

func InitB2Upload() (b2.FileInfo, error) {
	return service.B2.GetUploadURL(service.B2BucketID)
}

func PrepareUpload(
	id string,
	chunk int,
	data []byte,
) (FileUpload, db.B2Upload, error) {
	_, checksum := utils.GenChecksum(data)
	db.UpdateChecksums(id, checksum)

	metadata, err := db.RetrieveMetadata(id)
	if err != nil {
		return FileUpload{}, db.B2Upload{}, err
	}

	b2Values := db.GetB2UploadValues(id)

	upload := FileUpload{
		data:     data,
		filename: metadata.Name,
		salt:     metadata.Salt,
		checksum: checksum,
		chunk:    chunk,
		chunks:   metadata.Chunks,
	}

	return upload, b2Values, nil
}

func (upload FileUpload) Upload(b2Values db.B2Upload) (bool, error) {
	var err error

	if upload.chunks > 1 {
		largeFile := b2.FilePartInfo{
			FileID:             b2Values.UploadID,
			AuthorizationToken: b2Values.Token,
			UploadURL:          b2Values.UploadURL,
		}

		err = b2.UploadFilePart(
			largeFile,
			upload.chunk,
			upload.checksum,
			upload.data)

		if err != nil {
			return false, err
		}

		if upload.chunk == upload.chunks {
			b2ID, length := FinishLargeB2Upload(
				b2Values.UploadID,
				b2Values.Checksums)
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

		resp, err := b2.UploadFile(
			file,
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
	init, err := service.B2.StartLargeFile(filename, service.B2BucketID)
	if err != nil {
		panic(err)
	}

	return service.B2.GetUploadPartURL(init)
}

func FinishLargeB2Upload(b2ID string, checksums []string) (string, int) {
	largeFile, err := service.B2.FinishLargeFile(b2ID, checksums)
	if err != nil {
		panic(err)
	}

	return largeFile.FileID, largeFile.ContentLength
}

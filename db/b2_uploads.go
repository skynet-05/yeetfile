package db

import (
	"github.com/lib/pq"
	"log"
)

type B2Upload struct {
	MetadataID string
	UploadURL  string
	Token      string
	UploadID   string
	Checksums  []string
}

func InsertNewUpload(id string) B2Upload {
	s := `INSERT INTO b2_uploads (metadata_id)
	      VALUES ($1)`
	_, err := db.Exec(s, id)
	if err != nil {
		panic(err)
	}

	return B2Upload{MetadataID: id}
}

func (b2Upload B2Upload) UpdateUploadValues(
	uploadURL string,
	token string,
	uploadID string,
) bool {
	s := `UPDATE b2_uploads
	      SET upload_url=$1, token=$2, upload_id=$3
	      WHERE metadata_id=$4`
	_, err := db.Exec(s, uploadURL, token, uploadID, b2Upload.MetadataID)
	if err != nil {
		panic(err)
	}

	return true
}

func UpdateChecksums(id string, checksum string) bool {
	s := `UPDATE b2_uploads
	      SET checksums = array_append(checksums,$1)
	      WHERE metadata_id=$2`
	_, err := db.Exec(s, checksum, id)
	if err != nil {
		panic(err)
	}

	return true
}

func GetB2UploadValues(id string) B2Upload {
	s := `SELECT *
	      FROM b2_uploads
	      WHERE metadata_id = $1`

	rows, err := db.Query(s, id)
	if err != nil {
		log.Fatalf("Error retrieving upload values: %v", err)
		return B2Upload{}
	}

	if rows.Next() {
		var metadataID string
		var uploadURL string
		var token string
		var uploadID string
		var checksums []string

		err = rows.Scan(
			&metadataID,
			&uploadURL,
			&token,
			&uploadID,
			pq.Array(&checksums))

		return B2Upload{
			MetadataID: metadataID,
			UploadURL:  uploadURL,
			Token:      token,
			UploadID:   uploadID,
			Checksums:  checksums,
		}
	}

	return B2Upload{}
}

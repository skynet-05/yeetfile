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
	Local      bool
}

func CreateNewUpload(id string, name string) B2Upload {
	s := `INSERT INTO b2_uploads (metadata_id, upload_id)
	      VALUES ($1, $2)`
	_, err := db.Exec(s, id, name)
	if err != nil {
		panic(err)
	}

	return B2Upload{MetadataID: id}
}

func (b2Upload B2Upload) UpdateUploadValues(
	uploadURL string,
	token string,
	uploadID string,
	local bool,
) bool {
	s := `UPDATE b2_uploads
	      SET upload_url=$1, 
	          token=$2, 
	          upload_id=CASE WHEN $3 = '' THEN upload_id ELSE $3 END, 
	          local=$4
	      WHERE metadata_id=$5`
	_, err := db.Exec(s, uploadURL, token, uploadID, local, b2Upload.MetadataID)
	if err != nil {
		log.Printf("Error updating b2 upload values: %v\n", err)
		return false
	}

	return true
}

func UpdateUploadID(id string, metadataID string) bool {
	s := `UPDATE b2_uploads SET upload_id=$1 WHERE metadata_id=$2`
	_, err := db.Exec(s, id, metadataID)
	if err != nil {
		log.Printf("Error updating b2 upload id: %v\n", err)
		return false
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
		var local bool

		err = rows.Scan(
			&metadataID,
			&uploadURL,
			&token,
			&uploadID,
			pq.Array(&checksums),
			&local)

		return B2Upload{
			MetadataID: metadataID,
			UploadURL:  uploadURL,
			Token:      token,
			UploadID:   uploadID,
			Checksums:  checksums,
			Local:      local,
		}
	}

	return B2Upload{}
}

func DeleteB2Uploads(id string) bool {
	s := `DELETE FROM b2_uploads
	      WHERE metadata_id = $1`
	_, err := db.Exec(s, id)
	if err != nil {
		return false
	}

	return true
}

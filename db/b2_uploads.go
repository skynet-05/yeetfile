package db

type B2Upload struct {
	MetadataID string
	UploadURL  string
	Token      string
	UploadID   string
}

func InsertNewUpload(id string) B2Upload {
	s := `INSERT INTO b2_uploads (metadata_id) VALUES ($1)`
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
	s := `UPDATE b2_uploads SET upload_url=$1, token=$2, upload_id=$3 WHERE metadata_id=$4`
	_, err := db.Exec(s, uploadURL, token, uploadID, b2Upload.MetadataID)
	if err != nil {
		panic(err)
	}

	return true
}

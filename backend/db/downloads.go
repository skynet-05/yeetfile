package db

import (
	"log"
	"time"
	"yeetfile/shared"
)

// InitDownload creates a new entry in the downloads table with a file's
// ID and the current user's ID, as well as the number of chunks in the
// file. Permissions must be checked before creating this entry.
func InitDownload(fileID, userID string, chunks int) (string, error) {
	var id string
	var err error
	id, err = getIDByFileAndUserID(fileID, userID)
	if err == nil && len(id) > 0 {
		err = resetDownload(id)
		return id, err
	}

	id = shared.GenRandomString(8)
	for downloadIDExists(id) {
		id = shared.GenRandomString(8)
	}

	s := `INSERT INTO downloads (
                   id,
                   file_id,
                   user_id,
                   total_chunks,
                   updated)
	      VALUES ($1, $2, $3, $4, $5)`

	_, err = db.Exec(s, id, fileID, userID, chunks, time.Now().UTC())
	return id, err
}

// GetDownload retrieves the true file ID for the
func GetDownload(id, userID string) (string, error) {
	var fileID string
	s := `SELECT file_id FROM downloads WHERE id=$1 AND user_id=$2`
	err := db.QueryRow(s, id, userID).Scan(&fileID)
	return fileID, err
}

// UpdateDownload increments the specified download's chunk column, deleting
// the entry if the current chunk matches the max number of chunks of a file
// (indicating that the file has been downloaded entirely).
func UpdateDownload(id string) error {
	s := `WITH updated AS (
	          UPDATE downloads
	          SET chunk = chunk + 1, updated=$2
	          WHERE id=$1
	      )
	      DELETE FROM downloads
	             WHERE chunk + 1 >= total_chunks
	             AND id=$1`

	_, err := db.Exec(s, id, time.Now().UTC())
	return err
}

// CleanUpDownloads removes all in-progress downloads that haven't been updated
// in over an hour, indicating that the download is no longer active.
func CleanUpDownloads() {
	s := `DELETE FROM downloads WHERE updated < $1`
	_, err := db.Exec(s, time.Now().UTC().Add(-time.Hour))
	if err != nil {
		log.Printf("Error cleaning up downloads: %v\n", err)
	}
}

func RemoveDownloadByFileID(fileID, userID string) error {
	s := `DELETE FROM downloads WHERE file_id=$1 AND user_id=$2`
	_, err := db.Exec(s, fileID, userID)
	return err
}

func getIDByFileAndUserID(fileID, userID string) (string, error) {
	var id string
	s := `SELECT id FROM downloads WHERE file_id=$1 AND user_id=$2`
	err := db.QueryRow(s, fileID, userID).Scan(&id)
	return id, err
}

func resetDownload(id string) error {
	s := `UPDATE downloads SET chunk=0 WHERE id=$1`
	_, err := db.Exec(s, id)
	return err
}

func downloadIDExists(id string) bool {
	count := 0
	s := `SELECT COUNT(*) FROM downloads WHERE id=$1`
	_ = db.QueryRow(s, id).Scan(&count)

	return count > 0
}

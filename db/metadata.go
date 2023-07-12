package db

import (
	"log"
	"yeetfile/utils"
)

type FileMetadata struct {
	ID     string
	Chunks int
	Name   string
	Salt   []byte
}

func InsertMetadata(chunks int, filename string, salt []byte) (string, error) {
	id := utils.GenRandomString(32)

	// Ensure the id isn't already being used in the table
	for MetadataIDExists(id) {
		id = utils.GenRandomString(32)
	}

	s := `INSERT INTO metadata (id, chunks, filename, salt) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(s, id, chunks, filename, salt)
	if err != nil {
		panic(err)
	}

	return id, nil
}

func MetadataIDExists(id string) bool {
	rows, err := db.Query(`SELECT * FROM metadata WHERE id = $1`, id)
	if err != nil {
		log.Fatalf("Error validating metadata id: %v", err)
		return true
	}

	// If any rows are returned, the id exists
	if rows.Next() {
		return true
	}

	return false
}

func RetrieveMetadata(id string) FileMetadata {
	rows, err := db.Query(`SELECT 1 FROM metadata WHERE id = $1`, id)
	if err != nil {
		log.Fatalf("Error retrieving metadata: %v", err)
		return FileMetadata{}
	}

	if rows.Next() {
		var fileID string
		var chunks int
		var filename string

		err = rows.Scan(&fileID, &chunks, &filename)

		return FileMetadata{ID: fileID, Chunks: chunks, Name: filename}
	}

	return FileMetadata{}
}

func UpdateB2Metadata(id string, b2ID string, length int) bool {
	s := `UPDATE metadata SET b2_id=$1, length=$2 WHERE id=$4`
	_, err := db.Exec(s, b2ID, length, id)
	if err != nil {
		panic(err)
	}

	return true
}

func DeleteMetadata(id string) bool {
	s := `DELETE FROM metadata WHERE id = $1`
	_, err := db.Exec(s, id)
	if err != nil {
		return false
	}

	return true
}

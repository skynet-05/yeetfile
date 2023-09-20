package db

import (
	"database/sql"
	"log"
	"yeetfile/utils"
)

type FileMetadata struct {
	ID     string
	Chunks int
	Name   string
	Salt   []byte
	B2ID   string
	Length int
}

func NewMetadata(chunks int, filename string, salt []byte) (string, error) {
	id := utils.GenRandomString(32)

	// Ensure the id isn't already being used in the table
	for MetadataIDExists(id) {
		id = utils.GenRandomString(32)
	}

	s := `INSERT INTO metadata
	      (id, chunks, filename, salt, b2_id, length)
	      VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(s, id, chunks, filename, salt, "", -1)
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
	s := `SELECT * FROM metadata WHERE id = $1`
	rows, err := db.Query(s, id)
	if err != nil {
		log.Fatalf("Error retrieving metadata: %v", err)
		return FileMetadata{}
	}

	if rows.Next() {
		return ParseMetadata(rows)
	}

	log.Fatalf("No metadata found for id: %s", id)
	return FileMetadata{}
}

func UpdateB2Metadata(id string, b2ID string, length int) bool {
	s := `UPDATE metadata
	      SET b2_id=$1, length=$2
	      WHERE id=$3`
	_, err := db.Exec(s, b2ID, length, id)
	if err != nil {
		panic(err)
	}

	return true
}

func ParseMetadata(rows *sql.Rows) FileMetadata {
	var id string
	var chunks int
	var name string
	var salt []byte
	var b2ID string
	var length int

	err := rows.Scan(&id, &chunks, &name, &salt, &b2ID, &length)

	if err != nil {
		panic(err)
		return FileMetadata{}
	}

	return FileMetadata{
		ID:     id,
		Chunks: chunks,
		Name:   name,
		Salt:   salt,
		B2ID:   b2ID,
		Length: length,
	}
}

func DeleteMetadata(id string) bool {
	s := `DELETE FROM metadata
	      WHERE id = $1`
	_, err := db.Exec(s, id)
	if err != nil {
		return false
	}

	return true
}

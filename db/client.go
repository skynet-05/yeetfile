package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"yeetfile/service"
	"yeetfile/utils"
)

var db *sql.DB

func init() {
	var (
		host     = utils.GetEnvVar("YEETFILE_DB_HOST", "localhost")
		port     = utils.GetEnvVar("YEETFILE_DB_PORT", "5432")
		user     = utils.GetEnvVar("YEETFILE_DB_USER", "postgres")
		password = utils.GetEnvVar("YEETFILE_DB_PASS", "")
		dbname   = utils.GetEnvVar("YEETFILE_DB_NAME", "yeetfile")
	)

	connStr := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
}

func DeleteFileByID(id string) {
	metadata := RetrieveMetadata(id)

	// File must be deleted from B2 before removing from the database
	if service.B2.DeleteFile(metadata.B2ID, metadata.Name) {
		log.Printf("%s deleted from B2\n", metadata.ID)

		if DeleteMetadata(id) {
			log.Printf("%s metadata deleted\n",
				metadata.ID)
		} else {
			log.Printf("Failed to delete metadata for %s\n",
				metadata.ID)
		}

		if DeleteB2Uploads(id) {
			log.Printf("%s B2 info deleted\n",
				metadata.ID)
		} else {
			log.Printf("Failed to delete B2 info for %s\n",
				metadata.ID)
		}

		if DeleteExpiry(id) {
			log.Printf("%s expiry fields deleted\n",
				metadata.ID)
		} else {
			log.Printf("Failed to delete expiry fields for %s\n",
				metadata.ID)
		}
	} else {
		log.Printf("Failed to delete B2 file (metadata id: %s)\n",
			metadata.ID)
	}
}

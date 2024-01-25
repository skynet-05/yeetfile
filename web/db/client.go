package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"path/filepath"
	"yeetfile/web/service"
	"yeetfile/web/utils"
)

var db *sql.DB

func init() {
	var (
		host     = utils.GetEnvVar("YEETFILE_DB_HOST", "localhost")
		port     = utils.GetEnvVar("YEETFILE_DB_PORT", "5432")
		user     = utils.GetEnvVar("YEETFILE_DB_USER", "postgres")
		password = utils.GetEnvVar("YEETFILE_DB_PASS", "")
		dbname   = utils.GetEnvVar("YEETFILE_DB_NAME", "postgres")
	)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname)

	// Open db connection
	var err error
	db, err = sql.Open("postgres", connStr)

	if err != nil || db.Ping() != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Init db contents from scripts/init.sql
	log.Printf("Initializing DB...")
	path := filepath.Join("web", "db", "scripts", "init.sql")
	c, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error initializing database -- have you generated " +
			"the sql script (web/db/scripts/init.sh)?")
		panic(err)
	}

	sqlScript := string(c)
	_, err = db.Exec(sqlScript)
	if err != nil {
		panic(err)
	}
}

func clearDatabase(id string) {
	if DeleteMetadata(id) {
		log.Printf("%s metadata deleted\n", id)
	} else {
		log.Printf("Failed to delete metadata for %s\n", id)
	}

	if DeleteB2Uploads(id) {
		log.Printf("%s B2 info deleted\n", id)
	} else {
		log.Printf("Failed to delete B2 info for %s\n", id)
	}

	if DeleteExpiry(id) {
		log.Printf("%s expiry fields deleted\n", id)
	} else {
		log.Printf("Failed to delete expiry fields for %s\n", id)
	}
}

func DeleteFileByID(id string) {
	metadata := RetrieveMetadata(id)

	// File must be deleted from B2 before removing from the database
	if service.B2.DeleteFile(metadata.B2ID, metadata.Name) {
		log.Printf("%s deleted from B2\n", metadata.ID)

		clearDatabase(metadata.ID)
	} else {
		if len(metadata.B2ID) == 0 {
			clearDatabase(metadata.ID)
		} else {
			log.Printf("Failed to delete B2 file (id: %s, "+
				"metadata id: %s)\n",
				metadata.B2ID, metadata.ID)
		}
	}
}

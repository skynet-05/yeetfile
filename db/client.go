package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
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

func DeleteAllByID(id string) bool {
	metadataDeleted := DeleteMetadata(id)
	b2InfoDeleted := DeleteB2Uploads(id)
	expiryDeleted := DeleteExpiry(id)

	return metadataDeleted && b2InfoDeleted && expiryDeleted
}

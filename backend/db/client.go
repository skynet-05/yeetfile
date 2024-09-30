package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"yeetfile/backend/cache"
	"yeetfile/backend/service"
	"yeetfile/backend/utils"
)

var db *sql.DB

//go:embed scripts/create_tables.sql
var createTablesSQL string

func init() {
	var (
		host     = utils.GetEnvVar("YEETFILE_DB_HOST", "localhost")
		port     = utils.GetEnvVar("YEETFILE_DB_PORT", "5432")
		user     = utils.GetEnvVar("YEETFILE_DB_USER", "postgres")
		password = utils.GetEnvVar("YEETFILE_DB_PASS", "")
		dbname   = utils.GetEnvVar("YEETFILE_DB_NAME", "yeetfile")
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

	ping := db.Ping()
	if err != nil || ping != nil {
		log.Fatalf("Unable to connect to database!\n"+
			"Error: %v\nPing: %v\n", err, ping)
	}

	// Init db contents from scripts/init.sql
	log.Printf("Setting up DB tables...")
	_, err = db.Exec(createTablesSQL)
	if err != nil {
		log.Fatalf("Unable to initialize database!\n"+
			"--Error: %v\n", err)
	}
}

// clearDatabase removes all instances of a file ID from all tables in the database
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

// TableIDExists checks any table to see if the `id` column already has a value
// matching the provided id param
func TableIDExists(tableName, id string) bool {
	rows, err := db.Query(`SELECT * FROM `+tableName+` WHERE id=$1`, id)
	if err != nil {
		utils.Logf("Error checking for id in table '%s': %v", tableName, err)
		return true
	}

	// If any rows are returned, the id exists
	defer rows.Close()
	if rows.Next() {
		return true
	}

	return false
}

// DeleteFileByMetadata removes a file from B2 matching the provided file ID
func DeleteFileByMetadata(metadata FileMetadata) {
	log.Println("Deleting file by metadata (B2 errors are OK)")
	if err := cache.RemoveFile(metadata.ID); err != nil {
		log.Printf("Error removing cached file: %v\n", metadata.ID)
	} else {
		log.Printf("%s deleted from cache\n", metadata.ID)
	}

	if ok, err := service.B2.CancelLargeFile(metadata.B2ID); ok && err == nil {
		log.Printf("%s (large B2 upload) canceled\n", metadata.ID)
		clearDatabase(metadata.ID)
	} else if ok, err = service.B2.DeleteFile(metadata.B2ID, metadata.Name); ok && err == nil {
		log.Printf("%s deleted from B2\n", metadata.ID)
		clearDatabase(metadata.ID)
	} else {
		if len(metadata.B2ID) == 0 {
			clearDatabase(metadata.ID)
		} else {
			log.Printf("Failed to delete B2 file (id: %s, "+
				"metadata id: %s)\n",
				metadata.B2ID, metadata.ID)
			clearDatabase(metadata.ID)
		}
	}
}

func Close() {
	log.Println("Closing DB connection")
	err := db.Close()
	if err != nil {
		panic(err)
	}
}

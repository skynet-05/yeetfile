package db

import (
	"log"
	"time"
)

type FileExpiry struct {
	ID        string
	Downloads int
	Date      time.Time
}

func SetFileExpiry(id string, downloads int, date time.Time) {
	s := `INSERT INTO expiry
	      (id, downloads, date)
	      VALUES ($1, $2, $3)`
	_, err := db.Exec(s, id, downloads, date)
	if err != nil {
		panic(err)
	}
}

func DecrementDownloads(id string) int {
	s1 := `UPDATE expiry
	      SET downloads = downloads - 1
	      WHERE id=$1
	      AND downloads > 0`
	_, err := db.Exec(s1, id)
	if err != nil {
		panic(err)
	}

	s2 := `SELECT downloads FROM expiry WHERE id=$1 AND downloads >= 0`
	rows, err := db.Query(s2, id)

	if err != nil {
		log.Printf("Error retrieving download counter: %v\n", err)
		return -1
	}

	defer rows.Close()
	if rows.Next() {
		var downloads int
		err = rows.Scan(&downloads)

		if err == nil {
			return downloads
		}
	}

	return -1
}

func GetFileExpiry(metadataID string) FileExpiry {
	s := `SELECT * FROM expiry WHERE id=$1`
	rows, err := db.Query(s, metadataID)

	if err != nil {
		log.Printf("Error retrieving file expiry: %v\n", err)
		return FileExpiry{}
	}

	defer rows.Close()
	if rows.Next() {
		var id string
		var downloads int
		var date time.Time

		err = rows.Scan(&id, &downloads, &date)

		return FileExpiry{
			ID:        id,
			Downloads: downloads,
			Date:      date,
		}
	}

	return FileExpiry{}
}

func DeleteExpiry(id string) bool {
	s := `DELETE FROM expiry
	      WHERE id = $1`
	_, err := db.Exec(s, id)
	if err != nil {
		return false
	}

	return true
}

// CheckExpiry inspects each entry in the expiry table to see if a file's
// expiration date has been surpassed. If it has, the file is deleted. Runs
// recursively once per second and should be called in a background thread.
func CheckExpiry() {
	s := `SELECT id, date FROM expiry`
	rows, err := db.Query(s)

	if err != nil {
		log.Printf("Error retrieving file expiry: %v\n", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var id string
		var date time.Time

		err = rows.Scan(&id, &date)

		if err != nil {
			log.Printf("Error scanning rows: %v\n", err)
			return
		}

		if time.Now().UTC().After(date.UTC()) {
			// File has expired, remove from the DB and B2
			log.Printf("%s has expired, removing now\n", id)
			metadata, err := RetrieveMetadata(id)
			if err != nil {
				log.Printf("Metadata not found for id: " + id)
			} else {
				DeleteFileByMetadata(metadata)
			}
		}
	}

	time.Sleep(1 * time.Second)
	CheckExpiry()
}

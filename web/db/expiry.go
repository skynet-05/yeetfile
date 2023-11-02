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
		log.Fatalf("Error retrieving download counter: %v", err)
		return -1
	}

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
		log.Fatalf("Error retrieving file expiry: %v", err)
		return FileExpiry{}
	}

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

func CheckExpiry() {
	s := `SELECT id, date FROM expiry`
	rows, err := db.Query(s)
	if err != nil {
		log.Fatalf("Error retrieving file expiry: %v", err)
		return
	}

	for rows.Next() {
		var id string
		var date time.Time

		err = rows.Scan(&id, &date)

		if err != nil {
			log.Fatalf("Error scanning rows: %v", err)
			return
		}

		if time.Now().UTC().After(date.UTC()) {
			// File has expired, remove from the DB and B2
			log.Printf("%s has expired, removing now\n", id)
			DeleteFileByID(id)
		}
	}

	time.Sleep(1 * time.Second)
	CheckExpiry()
}

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

func DeleteExpiry(id string) bool {
	s := `DELETE FROM expiry
	      WHERE id = $1`
	_, err := db.Exec(s, id)
	if err != nil {
		return false
	}

	return true
}

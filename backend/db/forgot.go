package db

import (
	"database/sql"
	"time"
)

func CanRequestPasswordHint(email string) (bool, error) {
	var requested time.Time
	s := `SELECT requested FROM forgot WHERE email=$1`
	err := db.QueryRow(s, email).Scan(&requested)

	if err == sql.ErrNoRows {
		return true, nil
	} else if err != nil {
		return false, err
	}

	if requested.Add(time.Hour).Before(time.Now().UTC()) {
		// If it's been an hour, it's ok to send another request
		return true, nil
	}

	return false, nil
}

func AddForgotEntry(email string) error {
	s := `INSERT INTO forgot (email, requested)
	      VALUES ($1, $2)
	      ON CONFLICT (email) 
	      DO UPDATE SET requested = $2;`
	_, err := db.Exec(s, email, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

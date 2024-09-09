package db

import (
	"database/sql"
	"errors"
	"time"
	"yeetfile/shared"
)

var ChangeEmailEntryTooNew = errors.New("change email request already submitted")

// NewChangeEmailEntry adds an entry to the change_email table containing the
// user's ID and "old" email (the one they're replacing with a new email).
func NewChangeEmailEntry(userID, oldEmail string) (string, error) {
	id, date, err := getChangeInfoByUserID(userID)
	if err == nil && len(id) > 0 {
		if date.Add(time.Hour).After(time.Now().UTC()) {
			return id, ChangeEmailEntryTooNew
		} else {
			s := `UPDATE change_email SET date=$2 WHERE id=$1`
			_, err = db.Exec(s, id, time.Now().UTC())
		}
		return id, err
	}

	id = shared.GenRandomString(16)
	exists, err := changeEmailIDEntryExists(id)
	if err != nil {
		return "", err
	}

	for exists && err == nil {
		id = shared.GenRandomString(16)
		exists, err = changeEmailIDEntryExists(id)
		if err != nil {
			return "", err
		}
	}

	if len(oldEmail) == 0 {
		// Email must be unique in the table, but for account ID-only
		// users the email is blank. Setting it to the same value as the
		// ID works well enough for this.
		oldEmail = id
	}

	s := `INSERT INTO change_email 
	          (id, account_id, old_email, date) 
	      VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(s, id, userID, oldEmail, time.Now().UTC())
	return id, err
}

// IsChangeIDValid checks that the current user's ID matches the change ID that
// they're requesting
func IsChangeIDValid(changeID, userID string) bool {
	count := 0
	s := `SELECT COUNT(*) FROM change_email WHERE id=$1 AND account_id=$2`
	err := db.QueryRow(s, changeID, userID).Scan(&count)
	if err == nil && count > 0 {
		return true
	}

	return false
}

func RemoveEmailChangeByChangeID(changeID string) error {
	s := `DELETE FROM change_email WHERE id=$1`
	_, err := db.Exec(s, changeID)
	return err
}

func changeEmailIDEntryExists(id string) (bool, error) {
	count := 0
	s := `SELECT COUNT(*) FROM change_email WHERE id=$1`
	err := db.QueryRow(s, id).Scan(&count)
	if err == sql.ErrNoRows || (err == nil && count == 0) {
		return false, nil
	}

	return true, err
}

func getChangeInfoByUserID(userID string) (string, time.Time, error) {
	var id string
	var date time.Time
	s := `SELECT id, date FROM change_email WHERE account_id=$1`
	err := db.QueryRow(s, userID).Scan(&id, &date)
	if err == sql.ErrNoRows || (err == nil && len(id) == 0) {
		return "", time.Time{}, nil
	}

	return id, date, err
}

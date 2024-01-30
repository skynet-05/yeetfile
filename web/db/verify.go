package db

import (
	"errors"
	"time"
	"yeetfile/web/utils"
)

// NewVerification creates a new verification entry for a user
func NewVerification(email string, pwHash []byte, reset bool) (string, error) {
	if !reset {
		r, e := db.Query(`SELECT * FROM users WHERE email = $1`, email)

		if e != nil {
			return "", e
		} else if r.Next() {
			return "", UserAlreadyExists
		}
	}

	// Generate verification code to be sent to the user's email
	code := utils.GenRandomNumbers(6)

	rows, err := db.Query(`SELECT * FROM verify WHERE email = $1`, email)
	if rows.Next() {
		// This user already has a verification entry -- update the
		// code before resending the email
		s := `UPDATE verify SET code = $1 WHERE email=$2`
		_, err = db.Exec(s, code, email)
		if err != nil {
			return "", err
		}
	} else {
		s := `INSERT INTO verify (email, code, date, pw_hash) VALUES ($1, $2, $3, $4)`
		_, err = db.Exec(s, email, code, time.Now(), pwHash)
		if err != nil {
			return "", err
		}
	}

	return code, nil
}

// VerifyUser verifies the user's email against the code stored in the `verify`
// table. If the code matches and a password hash is returned, a new user
// is created in the `users` table. Returns the ID of the new user (if created)
// and an error.
func VerifyUser(email string, code string) ([]byte, error) {
	rows, err := db.Query(`SELECT pw_hash FROM verify 
                               WHERE email = $1 AND code = $2`, email, code)
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		var pwHash []byte
		err = rows.Scan(&pwHash)
		if err != nil {
			return nil, err
		}

		return pwHash, nil
	}

	return nil, errors.New("unable to find user")
}

// DeleteVerification removes a verification entry from the table
func DeleteVerification(email string) error {
	s := `DELETE FROM verify
	      WHERE email = $1`
	_, err := db.Exec(s, email)
	if err != nil {
		return err
	}

	return nil
}

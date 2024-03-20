package db

import (
	"errors"
	"time"
	"yeetfile/shared"
)

// NewVerification creates a new verification entry for a user
func NewVerification(
	ident string,
	pwHash []byte,
	protectedKey []byte,
	reset bool,
) (string, error) {
	if !reset {
		r, e := db.Query(`SELECT * FROM users WHERE email = $1 OR id = $1`, ident)

		if e != nil {
			return "", e
		} else if r.Next() {
			return "", UserAlreadyExists
		}
	}

	// Generate verification code
	code := shared.GenRandomNumbers(6)

	rows, err := db.Query(`SELECT * FROM verify WHERE identity = $1`, ident)
	if rows.Next() {
		// This user already has a verification entry -- update the
		// code before resending the verification request
		s := `UPDATE verify SET code = $1 WHERE identity=$2`
		_, err = db.Exec(s, code, ident)
		if err != nil {
			return "", err
		}
	} else {
		s := `INSERT INTO verify (identity, code, date, pw_hash, protected_key) 
		      VALUES ($1, $2, $3, $4, $5)`
		_, err = db.Exec(s, ident, code, time.Now(), pwHash, protectedKey)
		if err != nil {
			return "", err
		}
	}

	return code, nil
}

// VerifyUser verifies the user's email against the code stored in the `verify`
// table. If the code matches the user's password hash and protected key are
// returned so that a new user can be added to the `users` table.
func VerifyUser(identity string, code string) ([]byte, []byte, error) {
	rows, err := db.Query(`SELECT pw_hash, protected_key FROM verify 
                               WHERE identity = $1 AND code = $2`, identity, code)
	if err != nil {
		return nil, nil, err
	}

	if rows.Next() {
		var pwHash []byte
		var protectedKey []byte
		err = rows.Scan(&pwHash, &protectedKey)
		if err != nil {
			return nil, nil, err
		}

		return pwHash, protectedKey, nil
	}

	return nil, nil, errors.New("unable to find user")
}

// DeleteVerification removes a verification entry from the table
func DeleteVerification(identity string) error {
	s := `DELETE FROM verify
	      WHERE identity = $1`
	_, err := db.Exec(s, identity)
	if err != nil {
		return err
	}

	return nil
}

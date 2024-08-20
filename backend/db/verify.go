package db

import (
	"database/sql"
	"errors"
	"time"
	"yeetfile/backend/config"
	"yeetfile/shared"
)

var VerificationCodeExistsError = errors.New("verification code already sent")

type NewAccountValues struct {
	PasswordHash  []byte
	ProtectedKey  []byte
	PublicKey     []byte
	RootFolderKey []byte
}

// NewVerification creates a new verification entry for a user
func NewVerification(
	signupData shared.Signup,
	pwHash []byte,
	reset bool,
) (string, error) {
	if !reset {
		r, e := db.Query(`SELECT * FROM users WHERE email = $1 OR id = $1`,
			signupData.Identifier)

		if e != nil {
			return "", e
		} else if r.Next() {
			return "", UserAlreadyExists
		}
	}

	// Generate verification code
	code := shared.GenRandomNumbers(6)

	rows, err := db.Query(`SELECT date FROM verify WHERE identity = $1`,
		signupData.Identifier)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var date time.Time
		err = rows.Scan(&date)
		if err != nil {
			return "", err
		} else if time.Now().UTC().Before(date) {
			// This user already has a verification entry, but it's
			// too early to send a new code. Update the other values
			// (in the password changed) and move on
			s := `UPDATE verify
			      SET pw_hash=$1, 
			          protected_key=$2, 
			          public_key=$3, 
			          root_folder_key=$4 
			      WHERE identity=$5`
			_, err = db.Exec(s,
				pwHash,
				signupData.ProtectedKey,
				signupData.PublicKey,
				signupData.RootFolderKey,
				signupData.Identifier)
			if err != nil {
				return "", err
			}

			return "", VerificationCodeExistsError
		} else {
			// This user already has a verification entry -- update the
			// code before resending the verification request
			s := `UPDATE verify 
			      SET 
			          code=$1,
			          pw_hash=$2,
			          date=$3,
			          protected_key=$4,
			          public_key=$5,
			          root_folder_key=$6
			      WHERE identity=$7`
			_, err = db.Exec(s,
				code,
				pwHash,
				time.Now().Add(5*time.Minute).UTC(),
				signupData.ProtectedKey,
				signupData.PublicKey,
				signupData.RootFolderKey,
				signupData.Identifier)
			if err != nil {
				return "", err
			}

			return code, nil
		}
	} else {
		s := `INSERT INTO verify (
                    identity,
                    code,
                    date,
                    pw_hash,
                    protected_key,
                    public_key,
                    root_folder_key) 
		      VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err = db.Exec(
			s,
			signupData.Identifier,
			code,
			time.Now().Add(5*time.Minute).UTC(),
			pwHash,
			signupData.ProtectedKey,
			signupData.PublicKey,
			signupData.RootFolderKey)
		if err != nil {
			return "", err
		}
	}

	return code, nil
}

// VerifyUser verifies the user's email against the code stored in the `verify`
// table. If the code matches the user's password hash and protected key are
// returned so that a new user can be added to the `users` table.
func VerifyUser(identity string, code string) (NewAccountValues, error) {
	var rows *sql.Rows
	var err error

	// Skip code verification in debug mode
	if config.IsDebugMode {
		rows, err = db.Query(`SELECT pw_hash, protected_key, public_key, root_folder_key FROM verify 
                               WHERE identity = $1`, identity)
	} else {
		rows, err = db.Query(`SELECT pw_hash, protected_key, public_key, root_folder_key FROM verify 
                               WHERE identity = $1 AND code = $2`, identity, code)
	}

	if err != nil {
		return NewAccountValues{}, err
	}

	defer rows.Close()
	if rows.Next() {
		var pwHash []byte
		var protectedKey []byte
		var publicKey []byte
		var rootFolderKey []byte
		err = rows.Scan(&pwHash, &protectedKey, &publicKey, &rootFolderKey)
		if err != nil {
			return NewAccountValues{}, err
		}

		return NewAccountValues{
			PasswordHash:  pwHash,
			ProtectedKey:  protectedKey,
			PublicKey:     publicKey,
			RootFolderKey: rootFolderKey,
		}, nil
	}

	return NewAccountValues{}, errors.New("unable to find user")
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

package db

import (
	"database/sql"
	"errors"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/crypto"
	"yeetfile/shared"
)

var VerificationCodeExistsError = errors.New("verification code already sent")

type VerifiedAccountValues struct {
	AccountID               string
	Email                   string
	PasswordHash            []byte
	ProtectedPrivateKey     []byte
	PublicKey               []byte
	ProtectedVaultFolderKey []byte
	PasswordHint            []byte
}

// NewVerification creates a new verification entry for a user. Account ID can
// be left empty for new user verification, otherwise should be provided if
// an existing user is verifying their new email.
func NewVerification(
	signupData shared.Signup,
	pwHash []byte,
	accountID string,
) (string, error) {
	if config.YeetFileConfig.MaxUserCount > 0 {
		count, err := GetUserCount()
		if err != nil {
			return "", err
		} else if count == config.YeetFileConfig.MaxUserCount {
			return "", UserLimitReached
		}
	}

	r, e := db.Query(`SELECT * FROM users WHERE email = $1 OR id = $1`,
		signupData.Identifier)

	if e != nil {
		return "", e
	} else if r.Next() {
		return "", UserAlreadyExists
	}

	// Generate verification code
	code := shared.GenRandomNumbers(6)

	rows, err := db.Query(`SELECT date FROM verify WHERE identity = $1`,
		signupData.Identifier)
	if err != nil {
		return "", err
	}

	var pwHintEncrypted []byte
	if len(signupData.PasswordHint) > 0 {
		pwHintEncrypted, err = crypto.Encrypt(signupData.PasswordHint)
		if err != nil {
			return "", err
		}
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
			// (in case the password changed)
			s := `UPDATE verify
			      SET pw_hash=$1, 
			          public_key=$2, 
			          protected_private_key=$3, 
			          protected_vault_folder_key=$4, 
			          pw_hint=$5,
			          account_id=$6
			      WHERE identity=$7`
			_, err = db.Exec(s,
				pwHash,
				signupData.PublicKey,
				signupData.ProtectedPrivateKey,
				signupData.ProtectedVaultFolderKey,
				pwHintEncrypted,
				accountID,
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
			          public_key=$4,
			          protected_private_key=$5,
			          protected_vault_folder_key=$6,
			          pw_hint=$7,
			          account_id=$8
			      WHERE identity=$9`
			_, err = db.Exec(s,
				code,
				pwHash,
				time.Now().Add(5*time.Minute).UTC(),
				signupData.PublicKey,
				signupData.ProtectedPrivateKey,
				signupData.ProtectedVaultFolderKey,
				pwHintEncrypted,
				accountID,
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
                    public_key,
                    protected_private_key,
                    protected_vault_folder_key,
                    account_id,
                    pw_hint) 
		      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		_, err = db.Exec(
			s,
			signupData.Identifier,
			code,
			time.Now().Add(5*time.Minute).UTC(),
			pwHash,
			signupData.PublicKey,
			signupData.ProtectedPrivateKey,
			signupData.ProtectedVaultFolderKey,
			accountID,
			pwHintEncrypted)
		if err != nil {
			return "", err
		}
	}

	return code, nil
}

// VerifyUser verifies the user's email against the code stored in the `verify`
// table. If the code matches the user's password hash and protected key are
// returned so that a new user can be added to the `users` table.
func VerifyUser(identity string, code string) (VerifiedAccountValues, error) {
	var (
		accountID               string
		pwHash                  []byte
		publicKey               []byte
		protectedPrivateKey     []byte
		protectedVaultFolderKey []byte
		encPwHint               []byte
	)

	s := `SELECT 
	          account_id,
	          pw_hash, 
	          public_key, 
	          protected_private_key, 
	          protected_vault_folder_key, 
	          pw_hint 
	      FROM verify WHERE identity = $1`

	var row *sql.Row

	// Add code verification if not in debug mode
	if !config.IsDebugMode {
		s += ` AND code=$2`
		row = db.QueryRow(s, identity, code)
	} else {
		row = db.QueryRow(s, identity)
	}

	err := row.Scan(
		&accountID,
		&pwHash,
		&publicKey,
		&protectedPrivateKey,
		&protectedVaultFolderKey,
		&encPwHint)

	if err != nil {
		return VerifiedAccountValues{}, err
	}

	return VerifiedAccountValues{
		Email:                   identity,
		AccountID:               accountID,
		PasswordHash:            pwHash,
		PublicKey:               publicKey,
		ProtectedPrivateKey:     protectedPrivateKey,
		ProtectedVaultFolderKey: protectedVaultFolderKey,
		PasswordHint:            encPwHint,
	}, nil
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

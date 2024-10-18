package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"yeetfile/backend/db"
)

var (
	Missing2FAErr = errors.New("TOTP code missing")
	Failed2FAErr  = errors.New("TOTP code failed")
)

// ValidateCredentials checks the provided key hash against the one stored in
// the database, and if there's a match, returns the user's true account ID.
func ValidateCredentials(
	identifier string,
	keyHash []byte,
	code string,
	validate2FA bool,
) (string, error) {
	var userID string
	var pwHash []byte
	var secret []byte
	var err error
	if strings.Contains(identifier, "@") {
		pwHash, secret, err = db.GetUserPasswordHashByEmail(identifier)
		if err != nil {
			return "", err
		}

		userID, err = db.GetUserIDByEmail(identifier)
		if err != nil {
			return "", err
		}
	} else {
		pwHash, secret, err = db.GetUserPasswordHashByID(identifier)
		if err != nil {
			return "", err
		}

		userID = identifier
	}

	err = bcrypt.CompareHashAndPassword(pwHash, keyHash)
	if err != nil {
		return "", err
	}

	if secret != nil && len(secret) > 0 && validate2FA {
		err = validateTOTP(secret, code, userID)
		if err != nil {
			return "", err
		}
	}

	return userID, nil
}

func createNewUser(values db.VerifiedAccountValues) (string, error) {
	var id string
	var err error

	// Create new user
	if len(values.AccountID) > 0 {
		id = values.AccountID
		_, err = db.NewUser(db.User{
			ID:                  values.AccountID,
			PublicKey:           values.PublicKey,
			ProtectedPrivateKey: values.ProtectedPrivateKey,
			PasswordHash:        values.PasswordHash,
		})
	} else {
		id, err = db.NewUser(db.User{
			Email:               values.Email,
			PasswordHash:        values.PasswordHash,
			PublicKey:           values.PublicKey,
			ProtectedPrivateKey: values.ProtectedPrivateKey,
			PasswordHint:        values.PasswordHint,
		})
	}

	if err != nil {
		log.Printf("Error initializing new account: %v\n", err)
		return "", err
	}

	// Initialize user's root vault folder
	err = db.NewRootFolder(id, values.ProtectedVaultFolderKey)
	if err != nil {
		log.Printf("Error initializing user vault: %v\n", err)
		return "", err
	}

	// Initialize user pass metadata index
	err = db.InitPassIndex(id)
	if err != nil {
		log.Printf("Error initializing password index: %v\n", err)
		return "", err
	}

	return id, nil
}

func updateUser(values db.VerifiedAccountValues) error {
	err := db.UpdateUser(db.User{
		Email:               values.Email,
		PasswordHash:        values.PasswordHash,
		ProtectedPrivateKey: values.ProtectedPrivateKey,
	}, values.AccountID)

	if err != nil {
		return err
	}

	return nil
}

package auth

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"yeetfile/backend/db"
)

// ValidateCredentials checks the provided key hash against the one stored in
// the database, and if there's a match, returns the user's true account ID.
func ValidateCredentials(identifier string, keyHash []byte) (string, error) {
	var userID string
	var pwHash []byte
	var err error
	if strings.Contains(identifier, "@") {
		pwHash, err = db.GetUserPasswordHashByEmail(identifier)
		if err != nil {
			return "", err
		}

		userID, err = db.GetUserIDByEmail(identifier)
		if err != nil {
			return "", err
		}
	} else {
		pwHash, err = db.GetUserPasswordHashByID(identifier)
		if err != nil {
			return "", err
		}

		userID = identifier
	}

	err = bcrypt.CompareHashAndPassword(pwHash, keyHash)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func createNewUser(values db.VerifiedAccountValues) error {
	// Create new user
	id, err := db.NewUser(db.User{
		Email:        values.Email,
		PasswordHash: values.PasswordHash,
		ProtectedKey: values.ProtectedKey,
		PublicKey:    values.PublicKey,
		PasswordHint: values.PasswordHint,
	})

	if err != nil {
		log.Printf("Error initializing new account: %v\n", err)
		return err
	}

	err = db.NewRootFolder(id, values.RootFolderKey)
	if err != nil {
		log.Printf("Error initializing user vault: %v\n", err)
		return err
	}

	return nil
}

func updateUser(values db.VerifiedAccountValues) error {
	err := db.UpdateUser(db.User{
		Email:        values.Email,
		PasswordHash: values.PasswordHash,
		ProtectedKey: values.ProtectedKey,
	}, values.AccountID)

	if err != nil {
		return err
	}

	return nil
}

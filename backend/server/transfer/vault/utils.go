package vault

import (
	"errors"
	"log"
	"yeetfile/backend/db"
)

var OutOfSpaceError = errors.New("not enough storage available")

// CanUserUpload checks if the user has enough storage available to upload a
// file of the specified size
func CanUserUpload(size int64, id string) error {
	// Validate that the user has enough space to upload this file
	usedStorage, availableStorage, err := db.GetUserStorageLimits(id)
	if err != nil {
		log.Printf("Error validating ability to upload: %v\n", err)
		return err
	} else if availableStorage-usedStorage < size {
		return OutOfSpaceError
	}

	return nil
}

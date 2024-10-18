package vault

import (
	"errors"
	"log"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
)

var OutOfSpaceError = errors.New("not enough storage available")

// CanUserUpload checks if the user has enough storage available to upload a
// file of the specified size
func CanUserUpload(size int64, userID, folderID string) error {
	var (
		usedStorage      int64
		availableStorage int64
		err              error
	)

	// Skip check if storage limits aren't configured
	if config.YeetFileConfig.DefaultUserStorage < 0 {
		return nil
	}

	// Validate that the user has enough space to upload this file
	if len(folderID) == 0 || folderID == userID {
		usedStorage, availableStorage, err = db.GetUserStorageLimits(userID)
	} else {
		usedStorage, availableStorage, err = db.GetFolderOwnerStorage(folderID)
	}

	if err != nil {
		log.Printf("Error validating ability to upload: %v\n", err)
		return err
	} else if availableStorage-usedStorage < size {
		return OutOfSpaceError
	}

	return nil
}

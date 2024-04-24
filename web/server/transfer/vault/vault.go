package vault

import (
	"errors"
	"log"
	"strings"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/service"
)

func updateVaultFile(id, userID string, mod shared.ModifyVaultFile) error {
	if len(mod.Name) > 0 {
		err := db.UpdateVaultFileName(id, userID, mod.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateVaultFolder(id, userID string, mod shared.ModifyVaultFolder) error {
	if len(mod.Name) > 0 {
		err := db.UpdateVaultFolderName(id, userID, mod.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func shareVaultItem(
	share shared.ShareItemRequest,
	itemID string,
	userID string,
	isFolder bool,
) (shared.ShareInfo, error) {
	var err error
	recipientID := share.User
	if strings.Contains(share.User, "@") {
		recipientID, err = db.GetUserIDByEmail(share.User)
		if err != nil {
			return shared.ShareInfo{}, err
		}
	} else {
		_, err = db.GetUserByID(share.User)
		if err != nil {
			return shared.ShareInfo{}, err
		}
	}

	userName, _ := db.GetUserPublicName(userID)

	newShare := shared.NewSharedItem{
		ItemID:       itemID,
		UserID:       userID,
		SharerName:   userName,
		RecipientID:  recipientID,
		ProtectedKey: share.ProtectedKey,
		CanModify:    share.CanModify,
	}

	var shareID string
	var shareErr error
	if isFolder {
		shareID, shareErr = db.ShareFolder(newShare)
	} else {
		shareID, shareErr = db.ShareFile(newShare)
	}

	return shared.ShareInfo{ID: shareID, CanModify: share.CanModify}, shareErr
}

func deleteVaultFolder(id, userID string, isShared bool) error {
	if isShared {
		// Delete shared folder reference and return
		return db.DeleteSharedFolder(id, userID)
	}

	subfolders, err := db.GetSubfolders(id, userID)
	if err != nil {
		return err
	}

	for _, sub := range subfolders {
		err = deleteVaultFolder(sub.ID, userID, isShared)
		if err != nil {
			return err
		}
	}

	items, err := db.GetVaultItems(userID, id)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = deleteVaultFile(item.ID, userID, false)
		if err != nil {
			return err
		}
	}

	err = db.DeleteVaultFolder(id, userID)
	if err != nil {
		return err
	}

	return nil
}

func deleteVaultFile(id, userID string, isShared bool) (int, error) {
	if isShared {
		// Delete shared file reference and return
		return 0, db.DeleteSharedFile(id, userID)
	}

	metadata, err := db.RetrieveVaultMetadata(id, userID)
	if err != nil {
		return 0, err
	}

	b2Info := db.GetB2UploadValues(metadata.ID)
	if !service.B2.DeleteFile(metadata.B2ID, b2Info.Name) {
		log.Printf("Failed to delete vault file from b2: '%s'", metadata.B2ID)
		//return errors.New("failed to delete")
	}

	if !db.DeleteB2Uploads(metadata.ID) {
		log.Printf("Failed to delete b2 records for vault file: '%s'", metadata.ID)
		//return errors.New("failed to delete b2 records")
	}

	vaultDeleteErr := db.DeleteVaultFile(id, userID)
	if vaultDeleteErr != nil {
		log.Printf("Failed to delete vault file from database: '%s' -- %v", id, vaultDeleteErr)
		return 0, errors.New("failed to delete")
	}

	totalUploadSize := metadata.Length - (shared.TotalOverhead * metadata.Chunks)
	err = db.UpdateStorageUsed(userID, -totalUploadSize)
	if err != nil {
		log.Printf("Failed to update storage for user: %v\n", err)
	}

	return totalUploadSize, nil
}

package vault

import (
	"errors"
	"log"
	"strings"
	"yeetfile/backend/db"
	"yeetfile/backend/service"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

func updateVaultFile(id, userID string, mod shared.ModifyVaultItem) error {
	if len(mod.Name) > 0 {
		err := db.UpdateVaultFileName(id, userID, mod.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateVaultFolder(id, userID string, mod shared.ModifyVaultItem) error {
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
	var userName string
	recipientID := share.User
	if strings.Contains(share.User, "@") {
		recipientID, err = db.GetUserIDByEmail(share.User)
		if err != nil {
			return shared.ShareInfo{}, err
		}
		userName = share.User
	} else {
		_, err = db.GetUserByID(share.User)
		if err != nil {
			return shared.ShareInfo{}, err
		}
		userName = shared.FormatIDTail(share.User)
	}

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
		shareID, shareErr = db.ShareFolder(newShare, userID)
	} else {
		shareID, shareErr = db.ShareFile(newShare, userID)
	}

	return shared.ShareInfo{
		ID:        shareID,
		Recipient: userName,
		CanModify: share.CanModify,
	}, shareErr
}

func DeleteVaultFolder(id, userID string, isShared bool) error {
	if isShared {
		// Delete shared folder reference and return
		return db.DeleteSharedFolder(id, userID)
	}

	subfolders, err := db.GetSubfolders(id, userID, shared.FolderOwnershipInfo{})
	if err != nil {
		return err
	}

	for _, sub := range subfolders {
		err = DeleteVaultFolder(sub.ID, userID, isShared)
		if err != nil {
			return err
		}
	}

	items, _, err := db.GetVaultItems(userID, id)
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

	if len(metadata.B2ID) > 0 {
		b2Info := db.GetB2UploadValues(metadata.ID)
		if !service.B2.DeleteFile(metadata.B2ID, b2Info.Name) {
			log.Printf(
				"Unable to delete vault file from b2: '%s'",
				metadata.B2ID)
		}

		if !db.DeleteB2Uploads(metadata.ID) {
			log.Printf(
				"Failed to delete b2 records for vault file: '%s'",
				metadata.ID)
		}
	}

	vaultDeleteErr := db.DeleteVaultFile(id, userID)
	if vaultDeleteErr != nil {
		log.Printf("Failed to delete vault file from database: '%s' -- %v", id, vaultDeleteErr)
		return 0, errors.New("failed to delete")
	}

	totalUploadSize := metadata.Length - (constants.TotalOverhead * metadata.Chunks)
	err = db.UpdateStorageUsed(userID, -totalUploadSize)
	if err != nil {
		log.Printf("Failed to update storage for user: %v\n", err)
	}

	return totalUploadSize, nil
}

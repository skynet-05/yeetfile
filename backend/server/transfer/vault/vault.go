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

// DeleteVaultFolder recursively deletes the folder matching the specified
// folder ID and all of its subfolders, returning the amount of freed space
func DeleteVaultFolder(id, userID string, isShared bool) (int64, error) {
	freed := int64(0)
	if isShared {
		// Delete shared folder reference and return
		return 0, db.DeleteSharedFolder(id, userID)
	}

	subfolders, err := db.GetSubfolders(id, userID, shared.FolderOwnershipInfo{})
	if err != nil {
		return 0, err
	}

	for _, sub := range subfolders {
		subFreed, err := DeleteVaultFolder(sub.ID, userID, isShared)
		if err != nil {
			return 0, err
		}

		freed += subFreed
	}

	items, _, err := db.GetVaultItems(userID, id)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		freedBytes, err := deleteVaultFile(item.ID, userID, false)
		if err != nil {
			return 0, err
		}

		freed += freedBytes
	}

	err = db.DeleteVaultFolder(id, userID)
	if err != nil {
		return 0, err
	}

	err = db.RemoveShareEntryByItemID(id)

	return freed, err
}

// deleteVaultFile deletes the file matching the specified ID and the
func deleteVaultFile(id, userID string, isShared bool) (int64, error) {
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
		deleted, err := service.B2.DeleteFile(metadata.B2ID, b2Info.Name)
		if !deleted || err != nil {
			log.Printf(
				"Unable to delete vault file from b2: '%s'",
				metadata.B2ID)
		}
	}

	if !db.DeleteB2Uploads(metadata.ID) {
		log.Printf(
			"Failed to delete b2 records for vault file: '%s'",
			metadata.ID)
	}

	vaultDeleteErr := db.DeleteVaultFile(id, userID)
	if vaultDeleteErr != nil {
		log.Printf("Failed to delete vault file from database: '%s' -- %v", id, vaultDeleteErr)
		return 0, errors.New("failed to delete")
	}

	totalUploadSize := metadata.Length - int64(constants.TotalOverhead*metadata.Chunks)
	err = db.UpdateStorageUsed(userID, -totalUploadSize)
	if err != nil {
		log.Printf("Failed to update storage for user: %v\n", err)
	}

	_ = db.RemoveDownloadByFileID(id, userID)
	err = db.RemoveShareEntryByItemID(id)

	return totalUploadSize, err
}

func abortUpload(metadata db.FileMetadata, userID string, chunkLen int64, chunkNum int) {
	db.DeleteFileByMetadata(metadata)
	totalSize := chunkLen
	for chunkNum > 1 {
		totalSize += int64(constants.ChunkSize)
	}

	err := db.UpdateStorageUsed(userID, -totalSize)
	if err != nil {
		log.Printf("Error adjusting user storage during abort: %v\n", err)
	}
}

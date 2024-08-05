package db

import (
	"errors"
	"yeetfile/shared"
)

const sharingIDLength = 16

var AlreadySharedError = errors.New("item is already shared with this user")

// AddSharingEntry adds a new entry to the sharing table containing relevant
// info regarding a shared folder or file.
func AddSharingEntry(
	ownerID,
	recipientID,
	itemID string,
	isFolder,
	canModify bool,
) (string, error) {
	sharingID := shared.GenRandomString(sharingIDLength)
	for TableIDExists("sharing", sharingID) {
		sharingID = shared.GenRandomString(sharingIDLength)
	}

	s := `INSERT INTO sharing (id, owner_id, recipient_id, item_id, is_folder, can_modify) 
	      VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(s, sharingID, ownerID, recipientID, itemID, isFolder, canModify)
	return sharingID, err
}

// IsSharedWithRecipient checks to see if a file or folder has already been
// shared with a particular user
func IsSharedWithRecipient(ownerID, itemID, recipientID string) (bool, error) {
	s := `SELECT COUNT(*) FROM sharing WHERE owner_id=$1 AND item_id=$2 AND recipient_id=$3`
	rows, err := db.Query(s, ownerID, itemID, recipientID)
	if err != nil {
		return true, err
	}

	defer rows.Close()
	if rows.Next() {
		var count int
		err = rows.Scan(&count)
		if err != nil {
			return true, err
		}

		return count > 0, nil
	}

	return true, errors.New("no sharing entry found")
}

// ModifyShare updates records in the database related to a shared file or folder
func ModifyShare(ownerID string, shareEdit shared.ShareEdit, isFolder bool) error {
	err := UserCanEditItem(shareEdit.ItemID, ownerID, isFolder)
	if err != nil {
		return err
	}

	s := `WITH updated_sharing AS (
	        UPDATE sharing
	        SET can_modify=$1
	        WHERE id=$2 AND owner_id=$3
	        RETURNING recipient_id, item_id, can_modify, is_folder
	      )
	      SELECT recipient_id, item_id, can_modify, is_folder
	      FROM updated_sharing`

	rows, err := db.Query(s, shareEdit.CanModify, shareEdit.ID, ownerID)
	if err != nil {
		return err
	}

	defer rows.Close()
	if rows.Next() {
		var recipientID string
		var itemID string
		var canModify bool
		var folder bool

		err = rows.Scan(&recipientID, &itemID, &canModify, &folder)
		if err != nil {
			return err
		}

		if folder {
			err = ChangeFolderPermission(itemID, recipientID, canModify)
		} else {
			err = ChangeFilePermission(itemID, recipientID, canModify)
		}
	}

	return err
}

// RemoveShare removes a user's access to a shared folder
func RemoveShare(ownerID, itemID, shareID string, isFolder bool) error {
	err := UserCanEditItem(itemID, ownerID, isFolder)
	if err != nil {
		return err
	}

	s := `SELECT recipient_id FROM sharing WHERE id=$1 AND item_id=$2 AND owner_id=$3`
	rows, err := db.Query(s, shareID, itemID, ownerID)
	if err != nil {
		return err
	}

	defer rows.Close()
	if rows.Next() {
		var recipientID string
		err = rows.Scan(&recipientID)
		if err != nil {
			return err
		}

		if isFolder {
			err = DeleteSharedFolderByRefID(itemID, recipientID)
		} else {
			err = DeleteSharedFileByRefID(itemID, recipientID)
		}
	}

	err = RemoveShareEntry(shareID)
	return err
}

func RemoveShareEntry(id string) error {
	deleteQuery := `DELETE FROM sharing WHERE id=$1`
	_, err := db.Exec(deleteQuery, id)
	return err
}

func RemoveShareEntryByRecipient(recipientID, refID string) error {
	deleteQuery := `DELETE FROM sharing WHERE recipient_id=$1 AND item_id=$2`
	_, err := db.Exec(deleteQuery, recipientID, refID)
	return err
}

// GetShareInfo retrieves all records of how an item has been shared
func GetShareInfo(ownerID, itemID string, isFolder bool) ([]shared.ShareInfo, error) {
	err := UserCanEditItem(itemID, ownerID, isFolder)
	if err != nil {
		return nil, err
	}

	s := `SELECT id, recipient_id, can_modify 
	      FROM sharing 
	      WHERE owner_id=$1 AND item_id=$2 ORDER BY id`
	rows, err := db.Query(s, ownerID, itemID)
	if err != nil {
		return nil, err
	}

	var shareList []shared.ShareInfo
	defer rows.Close()
	for rows.Next() {
		var id string
		var recipientID string
		var canModify bool

		err = rows.Scan(&id, &recipientID, &canModify)
		if err != nil {
			return nil, err
		}

		name, err := GetUserPublicName(recipientID)
		if err != nil {
			name = "???"
		}

		shareList = append(shareList, shared.ShareInfo{
			ID:        id,
			Recipient: name,
			CanModify: canModify,
		})
	}

	return shareList, nil
}

// UserCanEditItem checks to see if a file or folder is editable by the current user
func UserCanEditItem(itemID, ownerID string, isFolder bool) error {
	if isFolder {
		ownership, err := GetFolderOwnership(itemID, ownerID)
		if err != nil {
			return err
		} else if !ownership.CanModify {
			return errors.New("user cannot modify this item")
		}
	} else {
		folderID, err := GetFileFolderID(itemID, ownerID)
		if err != nil {
			return err
		}

		if folderID == ownerID {
			ownership, err := GetFileOwnership(itemID, ownerID)
			if err != nil {
				return err
			} else if !ownership.CanModify {
				return errors.New("user cannot modify this file")
			}
		}

		ownership, err := CheckFolderOwnership(ownerID, folderID)
		if err != nil {
			return err
		} else if !ownership.CanModify {
			return errors.New("user cannot modify item")
		}

	}

	return nil
}

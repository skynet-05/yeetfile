package confirmation

import (
	"fmt"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/models"
)

func GenConfirmMsg(requestType internal.RequestType, item models.VaultItem) (string, string) {
	switch requestType {
	case internal.DeleteFileRequest:
		itemType := "file"
		if item.IsFolder {
			itemType = "folder"
		}

		return fmt.Sprintf("Are you sure you want to delete %s '%s'?",
			itemType, item.Name), "WARNING: This cannot be undone!"
	}

	return "", ""
}

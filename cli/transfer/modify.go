package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/requests"
	"yeetfile/shared"
)

// DeleteItem deletes either a file or folder from the user's vault
func DeleteItem(itemID string, isFolder bool) error {
	endpoint := GetModificationEndpoint(isFolder)
	url := endpoint.Format(config.UserConfig.Server, itemID)
	resp, err := requests.DeleteRequest(url)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("server error %d", resp.StatusCode)
		return errors.New(msg)
	}

	return nil
}

// RenameItem renames a file or folder to the user's specified new name. This
// name is encrypted and then encoded in hex before this function is called.
func RenameItem(itemID, hexEncName string, isFolder bool) error {
	endpoint := GetModificationEndpoint(isFolder)
	url := endpoint.Format(config.UserConfig.Server, itemID)

	var err error
	var reqData []byte
	if isFolder {
		rename := shared.ModifyVaultFolder{Name: hexEncName}
		reqData, err = json.Marshal(rename)
	} else {
		rename := shared.ModifyVaultFile{Name: hexEncName}
		reqData, err = json.Marshal(rename)
	}

	if err != nil {
		return err
	}

	resp, err := requests.PutRequest(url, reqData)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("server error %d", resp.StatusCode)
		return errors.New(msg)
	}

	return nil
}

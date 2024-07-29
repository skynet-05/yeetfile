package share

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/models"
	"yeetfile/cli/requests"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

const ReadPerm = "Read Only"
const WritePerm = "Read + Write"

type Action int

const (
	Cancel Action = iota
	Edit
	Remove
	Add
)

type Perm int

const (
	Read Perm = iota
	Write
)

func fetchSharedInfo(item models.VaultItem) ([]shared.ShareInfo, error) {
	endpoint := getEndpoint(item)
	url := endpoint.Format(config.UserConfig.Server, item.ID)
	resp, err := requests.GetRequest(url)
	if err != nil {
		return nil, err
	}

	var shares []shared.ShareInfo
	err = json.NewDecoder(resp.Body).Decode(&shares)
	if err != nil {
		return nil, err
	}

	return shares, nil
}

func removeAccess(
	item models.VaultItem,
	shares []shared.ShareInfo,
) ([]shared.ShareInfo, error) {
	var removed []shared.ShareInfo
	endpoint := getEndpoint(item)
	url := endpoint.Format(config.UserConfig.Server, item.ID)
	for _, share := range shares {
		deleteURL := fmt.Sprintf("%s?id=%s", url, share.ID)
		resp, err := requests.DeleteRequest(deleteURL)
		if err != nil {
			msg := fmt.Sprintf("Failed to remove %s -- %s",
				share.Recipient, err.Error())
			return removed, errors.New(msg)
		} else if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("Failed to remove %s -- "+
				"server error [%d]: %s",
				share.Recipient,
				resp.StatusCode,
				resp.Body)
			return removed, errors.New(msg)
		}

		removed = append(removed, share)
	}

	return removed, nil
}

func editPermissions(
	item models.VaultItem,
	shares []shared.ShareInfo,
) ([]shared.ShareInfo, error) {
	var updated []shared.ShareInfo
	endpoint := getEndpoint(item)
	url := endpoint.Format(config.UserConfig.Server, item.ID)
	for _, share := range shares {
		reqData, _ := json.Marshal(shared.ShareEdit{
			ID:        share.ID,
			ItemID:    item.ID,
			CanModify: share.CanModify,
		})

		resp, err := requests.PutRequest(url, reqData)
		if err != nil {
			return updated, err
		} else if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("server error [%d]: %s", resp.StatusCode, resp.Body)
			return updated, errors.New(msg)
		} else {
			updated = append(updated, share)
		}
	}

	return updated, nil
}

func shareItem(
	item models.VaultItem,
	decryptFunc crypto.CryptFunc,
	decryptKey []byte,
	recipient string,
	perm Perm,
) (shared.ShareInfo, error) {
	itemKey, err := decryptFunc(decryptKey, item.ProtectedKey)
	if err != nil {
		return shared.ShareInfo{}, err
	}

	userKey, err := generateUserProtectedKey(recipient, itemKey)
	if err != nil {
		return shared.ShareInfo{}, err
	}

	endpoint := getEndpoint(item)
	url := endpoint.Format(config.UserConfig.Server, item.ID)
	reqData, _ := json.Marshal(shared.ShareItemRequest{
		User:         recipient,
		CanModify:    perm == Write,
		ProtectedKey: userKey,
	})

	resp, err := requests.PostRequest(url, reqData)
	if err != nil {
		return shared.ShareInfo{}, err
	}

	var shareInfo shared.ShareInfo
	err = json.NewDecoder(resp.Body).Decode(&shareInfo)
	if err != nil {
		return shared.ShareInfo{}, err
	}

	return shareInfo, nil
}

func generateUserProtectedKey(
	recipient string,
	key []byte,
) ([]byte, error) {
	pubKeyURL := endpoints.PubKey.Format(config.UserConfig.Server)
	pubKeyURL = pubKeyURL + fmt.Sprintf("?user=%s", recipient)

	resp, err := requests.GetRequest(pubKeyURL)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		srvErr, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("[%d] %s", resp.StatusCode, srvErr)
		return nil, errors.New(msg)
	}

	var pubKeyResponse shared.PubKeyResponse
	err = json.NewDecoder(resp.Body).Decode(&pubKeyResponse)
	if err != nil {
		return nil, err
	}

	userItemKey, err := crypto.EncryptRSA(pubKeyResponse.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return userItemKey, nil
}

func getEndpoint(item models.VaultItem) endpoints.Endpoint {
	if item.IsFolder {
		return endpoints.ShareFolder
	} else {
		return endpoints.ShareFile
	}
}

package files

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/models"
	"yeetfile/cli/requests"
	"yeetfile/cli/transfer"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

var folderContexts = make(map[string]*VaultContext)

type VaultContext struct {
	FolderID string
	CanEdit  bool
	Crypto   crypto.CryptoCtx
	Folders  []shared.VaultFolder
	Files    []shared.VaultItem
	Content  []models.VaultItem
}

func FetchVaultContext(folderID string) (*VaultContext, error) {
	if context, ok := folderContexts[folderID]; ok {
		return context, nil
	}

	url := endpoints.VaultFolder.Format(config.UserConfig.Server, folderID)
	response, err := requests.GetRequest(url)
	if err != nil {
		log.Fatal(err)
	} else if response.StatusCode != http.StatusOK {
		return &VaultContext{}, errors.New(response.Status)
	}

	body, err := io.ReadAll(response.Body)

	var folderResp shared.VaultFolderResponse
	err = json.Unmarshal(body, &folderResp)
	if err != nil {
		return &VaultContext{}, err
	}

	cryptCtx, err := crypto.DeriveVaultCryptoContext(folderResp.KeySequence)
	if err != nil {
		return &VaultContext{}, err
	}

	ctx := VaultContext{
		FolderID: folderID,
		Crypto:   cryptCtx,
		Folders:  folderResp.Folders,
		Files:    folderResp.Items,
		CanEdit:  folderResp.CurrentFolder.CanModify,
	}

	folderContexts[folderID] = &ctx
	return &ctx, nil
}

// UploadFile uploads the file contained at the specified path to the user's
// vault in the current folder. Provides a progress callback to indicate how
// many chunks from the total have been uploaded.
func (ctx *VaultContext) UploadFile(path string, progress func(int, int)) error {
	file, stat, err := shared.GetFileInfo(path)
	if err != nil {
		return err
	}

	key, _ := crypto.GenerateRandomKey()
	protectedKey, err := ctx.Crypto.EncryptFunc(
		ctx.Crypto.EncryptionKey,
		key)
	if err != nil {
		return err
	}

	pending, err := transfer.InitVaultFile(
		file, stat, ctx.FolderID, protectedKey, key)
	if err != nil {
		return err
	}

	chunk := 0
	result, err := pending.UploadData(func() {
		chunk += 1
		progress(chunk, pending.NumChunks)
	})

	if err != nil {
		return err
	}

	totalSize := int(stat.Size()) + (constants.TotalOverhead * pending.NumChunks)
	ctx.InsertItem(models.VaultItem{
		ID:           result,
		Name:         utils.GetFilenameFromPath(path),
		IsFolder:     false,
		Size:         totalSize,
		Modified:     time.Now(),
		CanModify:    ctx.CanEdit,
		ProtectedKey: protectedKey,
	})

	return nil
}

func (ctx *VaultContext) CreateFolder(folderName string) error {
	key, _ := crypto.GenerateRandomKey()
	protectedKey, err := ctx.Crypto.EncryptFunc(
		ctx.Crypto.EncryptionKey,
		key)
	if err != nil {
		return err
	}

	response, err := transfer.CreateVaultFolder(
		folderName,
		ctx.FolderID,
		protectedKey,
		key)
	if err != nil {
		return err
	}

	ctx.InsertItem(models.VaultItem{
		ID:       response.ID,
		Name:     folderName,
		IsFolder: true,
		Modified: time.Now(),
	})

	return nil
}

func (ctx *VaultContext) Delete(item models.VaultItem) error {
	err := transfer.DeleteItem(item.ID, item.IsFolder)
	if err != nil {
		return err
	}

	ctx.removeItem(item.ID)
	return nil
}

func (ctx *VaultContext) Update(item models.VaultItem) {
	ctx.updateItem(item)
}

func (ctx *VaultContext) Rename(newName string, item models.VaultItem) error {
	key, err := ctx.Crypto.DecryptFunc(ctx.Crypto.DecryptionKey, item.ProtectedKey)
	if err != nil {
		return err
	}

	encName, err := crypto.EncryptChunk(key, []byte(newName))
	if err != nil {
		return err
	}

	hexEncName := hex.EncodeToString(encName)
	err = transfer.RenameItem(item.ID, hexEncName, item.IsFolder)
	if err != nil {
		return err
	}

	ctx.renameItem(item.ID, newName)
	return nil
}

func (ctx *VaultContext) Download(
	item models.VaultItem,
	progress func(int, int),
) (string, error) {
	key, err := ctx.Crypto.DecryptFunc(ctx.Crypto.DecryptionKey, item.ProtectedKey)
	if err != nil {
		return "", err
	}

	filename := item.Name
	_, statErr := os.Stat(filename)
	for statErr == nil {
		filename = shared.CreateNewSaveName(filename)
		_, statErr = os.Stat(filename)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return "", err
	}

	p, err := transfer.InitVaultDownload(item.ID, key, file)
	if err != nil {
		return "", err
	}

	chunks := 0
	err = p.DownloadData(func() {
		chunks += 1
		progress(chunks, p.NumChunks)
	})

	if err != nil {
		return "", err
	}

	return filename, nil
}

// InsertItem inserts a vault item into the current vault context
func (ctx *VaultContext) InsertItem(item models.VaultItem) {
	ctx.Content = append(ctx.Content, item)
	folderContexts[ctx.FolderID].Content = ctx.Content
}

func (ctx *VaultContext) removeItem(itemID string) {
	for i, item := range ctx.Content {
		if item.ID == itemID {
			ctx.Content[i] = ctx.Content[len(ctx.Content)-1]
			ctx.Content = ctx.Content[:len(ctx.Content)-1]
			return
		}
	}
}

func (ctx *VaultContext) renameItem(itemID, newName string) {
	for i, item := range ctx.Content {
		if item.ID == itemID {
			ctx.Content[i].Name = newName
			ctx.Content[i].Modified = time.Now()
			return
		}
	}
}

func (ctx *VaultContext) updateItem(updatedItem models.VaultItem) {
	for i, item := range ctx.Content {
		if item.ID == updatedItem.ID {
			ctx.Content[i] = updatedItem
			return
		}
	}
}

func (ctx *VaultContext) parseContent() ([]models.VaultItem, error) {
	var contents []models.VaultItem

	folders, err := ctx.parseFolders()
	if err != nil {
		return contents, nil
	}

	files, err := ctx.parseFiles()
	if err != nil {
		return contents, nil
	}

	contents = append(folders, files...)
	ctx.Content = contents
	return contents, nil
}

func (ctx *VaultContext) parseFolders() ([]models.VaultItem, error) {
	folderModels := []models.VaultItem{}
	for _, folder := range ctx.Folders {
		key, err := ctx.Crypto.DecryptFunc(
			ctx.Crypto.DecryptionKey,
			folder.ProtectedKey)
		if err != nil {
			return folderModels, err
		}

		nameBytes, _ := hex.DecodeString(folder.Name)
		name, _ := crypto.DecryptChunk(key, nameBytes)
		folderModels = append(folderModels, models.VaultItem{
			ID:           folder.ID,
			RefID:        folder.RefID,
			Name:         string(name),
			IsFolder:     true,
			Modified:     utils.LocalTimeFromUTC(folder.Modified),
			SharedWith:   folder.SharedWith,
			SharedBy:     folder.SharedBy,
			ProtectedKey: folder.ProtectedKey,
			IsOwner:      folder.IsOwner,
			CanModify:    folder.CanModify,
		})
	}

	return folderModels, nil
}

func (ctx *VaultContext) parseFiles() ([]models.VaultItem, error) {
	fileModels := []models.VaultItem{}
	for _, file := range ctx.Files {
		key, err := ctx.Crypto.DecryptFunc(
			ctx.Crypto.DecryptionKey,
			file.ProtectedKey)
		if err != nil {
			return fileModels, err
		}

		nameBytes, _ := hex.DecodeString(file.Name)
		name, _ := crypto.DecryptChunk(key, nameBytes)
		fileModels = append(fileModels, models.VaultItem{
			ID:           file.ID,
			RefID:        file.RefID,
			Name:         string(name),
			IsFolder:     false,
			Modified:     utils.LocalTimeFromUTC(file.Modified),
			Size:         file.Size,
			SharedWith:   file.SharedWith,
			SharedBy:     file.SharedBy,
			ProtectedKey: file.ProtectedKey,
			IsOwner:      file.IsOwner,
			CanModify:    file.CanModify,
		})
	}

	return fileModels, nil
}

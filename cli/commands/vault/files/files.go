package files

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/cli/transfer"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

var folderContexts = make(map[string]*VaultContext)

type VaultContext struct {
	FolderID string
	CanEdit  bool
	IsOwner  bool
	Crypto   crypto.CryptoCtx
	Folders  []shared.VaultFolder
	Files    []shared.VaultItem
	Content  []models.VaultItem
}

var keyPair crypto.KeyPair

func FetchVaultContext(folderID string, isPassVault bool) (*VaultContext, error) {
	if context, ok := folderContexts[folderID]; ok {
		return context, nil
	}

	folderResp, err := globals.API.FetchFolderContents(folderID, isPassVault)
	if err != nil {
		return &VaultContext{}, err
	}

	cryptCtx, err := keyPair.DeriveVaultCryptoContext(folderResp.KeySequence)
	if err != nil {
		return &VaultContext{}, err
	}

	ctx := VaultContext{
		FolderID: folderID,
		Crypto:   cryptCtx,
		Folders:  folderResp.Folders,
		Files:    folderResp.Items,
		CanEdit:  folderResp.CurrentFolder.CanModify,
		IsOwner:  folderResp.CurrentFolder.IsOwner,
	}

	folderContexts[folderID] = &ctx
	return &ctx, nil
}

// UploadFile uploads the file contained at the specified path to the user's
// vault in the current folder. Provides a progress callback to indicate how
// many chunks from the total have been uploaded. Returns the uploaded file
// size and any errors.
func (ctx *VaultContext) UploadFile(path string, progress func(int, int)) (int64, error) {
	file, stat, err := shared.GetFileInfo(path)
	if err != nil {
		return 0, err
	}

	key, _ := crypto.GenerateRandomKey()
	protectedKey, err := ctx.Crypto.EncryptFunc(ctx.Crypto.EncryptionKey, key)
	if err != nil {
		return 0, err
	}

	pending, err := transfer.InitVaultFile(
		file, stat, ctx.FolderID, protectedKey, key)
	if err != nil {
		return 0, err
	}

	chunk := 0
	result, err := pending.UploadData(func() {
		chunk += 1
		progress(chunk, pending.NumChunks)
	})

	if err != nil {
		return 0, err
	}

	totalSize := stat.Size() + int64(constants.TotalOverhead*pending.NumChunks)
	ctx.InsertItem(models.VaultItem{
		ID:           result,
		RefID:        result,
		Name:         utils.GetFilenameFromPath(path),
		IsFolder:     false,
		Size:         totalSize,
		Modified:     time.Now(),
		CanModify:    ctx.CanEdit,
		IsOwner:      ctx.IsOwner,
		ProtectedKey: protectedKey,
	})

	return stat.Size(), nil
}

func (ctx *VaultContext) UploadPassEntry(item models.VaultItem) error {
	key, _ := crypto.GenerateRandomKey()
	protectedKey, err := ctx.Crypto.EncryptFunc(ctx.Crypto.EncryptionKey, key)
	if err != nil {
		return err
	}

	encName, err := crypto.EncryptChunk(key, []byte(item.Name))
	if err != nil {
		return err
	}

	name := hex.EncodeToString(encName)

	jsonData, err := json.Marshal(item.PassEntry)
	if err != nil {
		return err
	}

	encPassData, err := crypto.EncryptChunk(key, jsonData)
	if err != nil {
		return err
	}

	upload := shared.VaultUpload{
		Name:         name,
		Length:       1,
		Chunks:       1,
		FolderID:     ctx.FolderID,
		ProtectedKey: protectedKey,
		PasswordData: encPassData,
	}

	meta, err := globals.API.InitVaultFile(upload)
	if err != nil {
		return err
	}

	ctx.InsertItem(models.VaultItem{
		ID:           meta.ID,
		RefID:        meta.ID,
		Name:         item.Name,
		IsFolder:     false,
		Size:         1,
		Modified:     time.Now(),
		CanModify:    ctx.CanEdit,
		IsOwner:      ctx.IsOwner,
		ProtectedKey: protectedKey,
		PassEntry:    item.PassEntry,
	})

	return nil
}

func (ctx *VaultContext) UpdatePassEntry(item models.VaultItem) error {
	key, err := ctx.Crypto.DecryptFunc(ctx.Crypto.DecryptionKey, item.ProtectedKey)
	if err != nil {
		return err
	}

	encName, err := crypto.EncryptChunk(key, []byte(item.Name))
	if err != nil {
		return err
	}

	hexEncName := hex.EncodeToString(encName)

	jsonData, err := json.Marshal(item.PassEntry)
	if err != nil {
		return err
	}

	encPassData, err := crypto.EncryptChunk(key, jsonData)
	if err != nil {
		return err
	}

	err = globals.API.ModifyVaultFile(item.RefID, shared.ModifyVaultItem{
		Name:         hexEncName,
		PasswordData: encPassData,
	})

	if err != nil {
		return err
	}

	ctx.updateItem(item)
	return nil
}

func (ctx *VaultContext) CreateFolder(folderName string, isPassVault bool) error {
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
		key,
		isPassVault)
	if err != nil {
		return err
	}

	ctx.InsertItem(models.VaultItem{
		ID:       response.ID,
		RefID:    response.ID,
		Name:     folderName,
		IsFolder: true,
		Modified: time.Now(),
	})

	return nil
}

func (ctx *VaultContext) Delete(item models.VaultItem) error {
	err := transfer.DeleteItem(item.ID, len(item.SharedBy) > 0, item.IsFolder)
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
	err = transfer.RenameItem(ctx.getItemID(item), hexEncName, item.IsFolder)
	if err != nil {
		return err
	}

	ctx.renameItem(ctx.getItemID(item), newName)
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

	p, err := transfer.InitVaultDownload(ctx.getItemID(item), key, file)
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
		if item.RefID == itemID {
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

		var passEntry shared.PassEntry
		if file.PasswordData != nil && len(file.PasswordData) > 0 {
			passEntryData, _ := crypto.DecryptChunk(key, file.PasswordData)
			err = json.Unmarshal(passEntryData, &passEntry)
		}

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
			PassEntry:    passEntry,
		})
	}

	return fileModels, nil
}

func (ctx *VaultContext) getItemID(item models.VaultItem) string {
	if len(ctx.FolderID) > 0 {
		return item.ID
	} else {
		return item.RefID
	}
}

package vault

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"yeetfile/backend/cache"
	"yeetfile/backend/db"
	"yeetfile/backend/server/session"
	"yeetfile/backend/server/transfer"
	"yeetfile/backend/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

// FileHandler directs all requests to the appropriate handler for interacting
// with YeetFile Vault files
func FileHandler(w http.ResponseWriter, req *http.Request, userID string) {
	var fn session.HandlerFunc
	switch req.Method {
	case http.MethodPut, http.MethodDelete:
		fn = ModifyFileHandler
	}

	fn(w, req, userID)
}

// FolderHandler directs all requests to the appropriate handler for interacting
// with YeetFile Vault folders
func FolderHandler(w http.ResponseWriter, req *http.Request, userID string) {
	var fn session.HandlerFunc
	switch req.Method {
	case http.MethodPut, http.MethodDelete:
		fn = ModifyFolderHandler
	case http.MethodPost:
		fn = NewFolderHandler
	case http.MethodGet:
		fn = FolderViewHandler
	}

	fn(w, req, userID)
}

// FolderViewHandler returns folder contents for the requested folder. If a
// folder ID wasn't included in the request, the user's root level folder
// (distinguished by having the same ID as their account) is returned.
func FolderViewHandler(w http.ResponseWriter, req *http.Request, userID string) {
	var folderID string
	segments := utils.GetTrailingURLSegments(endpoints.VaultFolder, req.URL.Path)
	if len(segments) == 0 || len(segments[0]) == 0 {
		folderID = userID
	} else {
		folderID = segments[0]
	}

	items, ownership, err := db.GetVaultItems(userID, folderID)
	if err != nil {
		utils.Logf("Error fetching vault items: %v\n", err)

		if err == db.AccessError {
			http.Error(w, "Unauthorized access",
				http.StatusForbidden)
		} else {
			http.Error(w, "Error fetching vault items",
				http.StatusInternalServerError)
		}

		return
	}

	folder, _ := db.GetFolderInfo(folderID, userID, ownership, false)
	folders, _ := db.GetSubfolders(folderID, userID, ownership)
	keySequence, _ := db.GetKeySequence(folderID, userID)

	_ = json.NewEncoder(w).Encode(shared.VaultFolderResponse{
		Items:         items,
		Folders:       folders,
		CurrentFolder: folder,
		KeySequence:   keySequence,
	})
}

// NewFolderHandler handles the creation of vault folders
func NewFolderHandler(w http.ResponseWriter, req *http.Request, userID string) {
	var folder shared.NewVaultFolder
	err := json.NewDecoder(req.Body).Decode(&folder)
	if err != nil {
		utils.Logf("Error decoding request body: %v\n", err)
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	folderID, err := db.NewFolder(folder, userID)
	if err != nil {
		http.Error(w, "Error creating new folder", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(shared.NewFolderResponse{ID: folderID})
	if err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
		return
	}
}

// ModifyFolderHandler receives request to change or delete an existing folder.
func ModifyFolderHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	idPart := strings.Split(segments[len(segments)-1], "?")
	id := idPart[0]

	isShared := len(req.URL.Query().Get("shared")) > 0

	var modErr error
	switch req.Method {
	case http.MethodPut:
		var folderMod shared.ModifyVaultItem
		modErr = json.NewDecoder(req.Body).Decode(&folderMod)
		if modErr != nil {
			break
		}

		modErr = updateVaultFolder(id, userID, folderMod)
		break
	case http.MethodDelete:
		modErr = DeleteVaultFolder(id, userID, isShared)
		break
	}

	if modErr != nil {
		http.Error(w, "Error modifying folder", http.StatusBadRequest)
	}
}

// ModifyFileHandler handles requests to modify an existing file in the user's vault
func ModifyFileHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	idPart := strings.Split(segments[len(segments)-1], "?")
	id := idPart[0]

	isShared := len(req.URL.Query().Get("shared")) > 0

	var modErr error
	var modResponse []byte
	switch req.Method {
	case http.MethodPut:
		var fileMod shared.ModifyVaultItem
		modErr = json.NewDecoder(req.Body).Decode(&fileMod)
		if modErr != nil {
			utils.Logf("Error updating file: %v\n", modErr)
			break
		}
		modErr = updateVaultFile(id, userID, fileMod)
		break
	case http.MethodDelete:
		var freed int
		freed, modErr = deleteVaultFile(id, userID, isShared)

		if modErr == nil {
			modResponse, _ = json.Marshal(shared.DeleteResponse{FreedSpace: freed})
		}
		break
	}

	if modErr != nil {
		log.Printf("Error modifying file: %v\n", modErr)
		http.Error(w, "Error modifying file", http.StatusBadRequest)
	} else if modResponse != nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(modResponse)
	}
}

// UploadMetadataHandler initializes a file in the user's vault
func UploadMetadataHandler(w http.ResponseWriter, req *http.Request, userID string) {
	var upload shared.VaultUpload
	err := json.NewDecoder(req.Body).Decode(&upload)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	err = CanUserUpload(upload.Length, userID)
	if err != nil {
		log.Printf("Error checking if user can upload file: %v\n", err)
		http.Error(w, "Not enough storage available", http.StatusBadRequest)
		return
	}

	itemID, err := db.AddVaultItem(userID, upload)
	if err != nil {
		log.Printf("Error initializing vault upload: %v\n", err)
		http.Error(w, "Error initializing vault upload", http.StatusBadRequest)
		return
	}

	b2Upload := db.CreateNewUpload(itemID, upload.Name)

	var b2Err error
	if upload.Chunks == 1 {
		b2Err = transfer.InitB2Upload(b2Upload)
	} else {
		b2Err = transfer.InitLargeB2Upload(upload.Name, b2Upload)
	}

	if b2Err != nil {
		http.Error(w, "Error initializing storage", http.StatusInternalServerError)
		_ = db.DeleteVaultFile(itemID, userID)
		return
	}

	err = json.NewEncoder(w).Encode(shared.MetadataUploadResponse{ID: itemID})
	if err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
		return
	}
}

// UploadDataHandler processes incoming chunks of encrypted file data for a
// vault file
func UploadDataHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-2]
	chunkNum, err := strconv.Atoi(segments[len(segments)-1])
	if err != nil {
		http.Error(w, "Invalid upload URL", http.StatusBadRequest)
		return
	}

	data, err := utils.LimitedReader(w, req.Body)
	if err != nil {
		utils.Logf("[YF Vault] Error reading uploaded data: %v\n", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	totalSize := len(data) - constants.TotalOverhead
	metadata, err := db.RetrieveVaultMetadata(id, userID)
	if err != nil {
		utils.Logf("[YF Vault] Error fetching metadata: %v\n", err)
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	} else if chunkNum > metadata.Chunks {
		utils.Logf("[YF Vault] User uploading beyond stated # of chunks")
		http.Error(w, "Attempting to upload more chunks than specified",
			http.StatusBadRequest)
		abortUpload(metadata, userID, totalSize, chunkNum)
		return
	}

	err = db.UpdateStorageUsed(userID, totalSize)
	if err != nil {
		abortUpload(metadata, userID, totalSize, chunkNum)
		http.Error(w, "Attempting to upload beyond max storage",
			http.StatusBadRequest)
		return
	}

	upload, b2Values, err := transfer.PrepareUpload(metadata, chunkNum, data)
	if err != nil {
		http.Error(w, "Unable to initialize chunk upload",
			http.StatusBadRequest)
		abortUpload(metadata, userID, totalSize, chunkNum)
		return
	}

	done, err := upload.Upload(b2Values)
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		log.Printf("[YF Vault] Error uploading file: %v\n", err)
		abortUpload(metadata, userID, totalSize, chunkNum)
		return
	}

	if done {
		_, _ = io.WriteString(w, id)
	}
}

// DownloadHandler handles incoming requests for metadata pertaining to a file
// in the vault that a user wants to download
func DownloadHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	metadata, err := db.RetrieveVaultMetadata(id, userID)
	if err != nil {
		utils.Logf("Error fetching metadata: %v\n", err)
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	}

	response := shared.VaultDownloadResponse{
		Name:         metadata.Name,
		ID:           metadata.RefID,
		Chunks:       metadata.Chunks,
		Size:         metadata.Length,
		ProtectedKey: metadata.ProtectedKey,
	}

	jsonData, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

// DownloadChunkHandler handles requests for encrypted file data for a file in
// the user's vault
func DownloadChunkHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")

	if len(segments) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := segments[len(segments)-2]
	chunk, _ := strconv.Atoi(segments[len(segments)-1])
	if chunk <= 0 {
		chunk = 1 // Downloads always begin with chunk 1
	}

	metadata, err := db.RetrieveVaultMetadata(id, userID)
	if err != nil {
		utils.Logf("Error fetching metadata: %v\n", err)
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	}

	var bytes []byte
	if cache.HasFile(id, metadata.Length) {
		_, bytes = transfer.DownloadFileFromCache(id, metadata.Length, chunk)
	} else {
		cache.PrepCache(id, metadata.Length)
		_, bytes = transfer.DownloadFile(metadata.B2ID, metadata.Length, chunk)
		_ = cache.Write(id, bytes)
	}

	_, _ = w.Write(bytes)
}

// ShareHandler handles requests to share files or folders within the user's
// vault, as well as modifying the shared state of those files/folders
func ShareHandler(isFolder bool) session.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, userID string) {
		segments := strings.Split(req.URL.Path, "/")
		itemID := segments[len(segments)-1]

		if len(itemID) != db.VaultIDLength {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		var shareErr error
		switch req.Method {
		case http.MethodPost:
			var share shared.ShareItemRequest
			err := json.NewDecoder(req.Body).Decode(&share)
			if err != nil {
				http.Error(w, "Error decoding request",
					http.StatusBadRequest)
				return
			}

			var shareInfo shared.ShareInfo
			shareInfo, shareErr = shareVaultItem(share, itemID, userID, isFolder)
			if shareErr == nil {
				jsonData, _ := json.Marshal(shareInfo)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(jsonData)
				return
			}
			break
		case http.MethodGet:
			shares, err := db.GetShareInfo(userID, itemID, isFolder)
			if err != nil {
				http.Error(w, "Error fetching share info", http.StatusInternalServerError)
			}

			jsonData, _ := json.Marshal(shares)

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jsonData)
			return
		case http.MethodPut:
			var edit shared.ShareEdit
			err := json.NewDecoder(req.Body).Decode(&edit)
			if err != nil {
				http.Error(w, "Error decoding request",
					http.StatusBadRequest)
				return
			}

			shareErr = db.ModifyShare(userID, edit, isFolder)
		case http.MethodDelete:
			shareID := req.URL.Query().Get("id")
			if len(shareID) == 0 {
				http.Error(w, "Missing 'id' param", http.StatusBadRequest)
				return
			}

			shareErr = db.RemoveShare(userID, itemID, shareID, isFolder)
		}

		if shareErr != nil {
			log.Printf("Error with shared content: %v\n", shareErr)
			http.Error(w, "Error with shared content", http.StatusBadRequest)
			return
		}
	}
}

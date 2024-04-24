package vault

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"yeetfile/shared"
	"yeetfile/web/cache"
	"yeetfile/web/db"
	"yeetfile/web/server/session"
	"yeetfile/web/server/transfer"
	"yeetfile/web/utils"
)

func FolderViewHandler(root bool) session.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, userID string) {
		folderID := userID
		if !root {
			segments := strings.Split(req.URL.Path, "/")
			folderID = segments[len(segments)-1]
			if len(folderID) == 0 {
				folderID = userID
			}
		}

		items, err := db.GetVaultItems(userID, folderID)
		if err != nil {
			utils.Logf("Error fetching vault items: %v\n", err)
			http.Error(w, "Error fetching vault items",
				http.StatusInternalServerError)
			return
		}

		folder, err := db.GetFolderInfo(folderID, userID, false)
		folders, err := db.GetSubfolders(folderID, userID)
		keySequence, err := db.GetKeySequence(folderID, userID)

		_ = json.NewEncoder(w).Encode(shared.VaultFolderResponse{
			Items:         items,
			Folders:       folders,
			CurrentFolder: folder,
			KeySequence:   keySequence,
		})
	}
}

func SharedFolderViewHandler(root bool) session.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, userID string) {
		folderID := ""
		if !root {
			segments := strings.Split(req.URL.Path, "/")
			folderID = segments[len(segments)-1]
			if len(folderID) == 0 {
				folderID = ""
			}
		}

		items, err := db.GetSharedItems(userID, folderID)
		if err != nil {
			utils.Logf("Error fetching shared vault items: %v\n", err)
			http.Error(w, "Error fetching shared vault items",
				http.StatusInternalServerError)
			return
		}

		folder, err := db.GetFolderInfo(folderID, userID, false)
		folders, err := db.GetSubfolders(folderID, userID)
		keySequence, err := db.GetKeySequence(folderID, userID)

		_ = json.NewEncoder(w).Encode(shared.VaultFolderResponse{
			Items:         items,
			Folders:       folders,
			CurrentFolder: folder,
			KeySequence:   keySequence,
		})
	}
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

func PublicFolderHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	var pubErr error
	switch req.Method {
	case http.MethodPost:
		var pubFolder shared.NewPublicVaultFolder
		pubErr = json.NewDecoder(req.Body).Decode(&pubFolder)
		if pubErr != nil {
			break
		}

		pubErr = db.NewPublicFolder(pubFolder, userID)
		break
	case http.MethodDelete:
		pubErr = db.DeletePublicFolder(id, userID)
		break
	}

	if pubErr != nil {
		http.Error(w, "Error modifying public folder", http.StatusBadRequest)
	}
}

func ModifyFolderHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	idPart := strings.Split(segments[len(segments)-1], "?")
	id := idPart[0]

	isShared := len(req.URL.Query().Get("shared")) > 0

	var modErr error
	switch req.Method {
	case http.MethodPut:
		var folderMod shared.ModifyVaultFolder
		modErr = json.NewDecoder(req.Body).Decode(&folderMod)
		if modErr != nil {
			break
		}

		modErr = updateVaultFolder(id, userID, folderMod)
		break
	case http.MethodDelete:
		modErr = deleteVaultFolder(id, userID, isShared)
		break
	}

	if modErr != nil {
		http.Error(w, "Error modifying folder", http.StatusBadRequest)
	}
}

func ModifyFileHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	idPart := strings.Split(segments[len(segments)-1], "?")
	id := idPart[0]

	isShared := len(req.URL.Query().Get("shared")) > 0

	var modErr error
	var modResponse []byte
	switch req.Method {
	case http.MethodPut:
		var fileMod shared.ModifyVaultFile
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
		http.Error(w, "Error modifying file", http.StatusBadRequest)
	} else if modResponse != nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(modResponse)
	}
}

func UploadMetadataHandler(w http.ResponseWriter, req *http.Request, userID string) {
	var upload shared.VaultUpload
	err := json.NewDecoder(req.Body).Decode(&upload)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
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
		return
	}

	err = json.NewEncoder(w).Encode(shared.VaultUploadResponse{ID: itemID})
	if err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
		return
	}
}

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
		utils.Logf("Error reading uploaded data: %v\n", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	metadata, err := db.RetrieveVaultMetadata(id, userID)
	if err != nil {
		utils.Logf("Error fetching metadata: %v\n", err)
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	} else if chunkNum > metadata.Chunks {
		utils.Logf("User uploading allocated number of chunks")
		http.Error(w, "Attempting to upload more chunks than specified", http.StatusBadRequest)
		return
	}

	upload, b2Values, err := transfer.PrepareUpload(metadata, chunkNum, data)

	done, err := upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	if done {
		var totalUploadSize int
		if metadata.Chunks == 1 {
			totalUploadSize = len(data) - shared.TotalOverhead
		} else {
			totalUploadSize = len(data) +
				(shared.ChunkSize * (metadata.Chunks - 1)) -
				(shared.TotalOverhead * (metadata.Chunks - 1))
		}

		err = db.UpdateStorageUsed(userID, totalUploadSize)
		_, _ = io.WriteString(w, id)
	}
}

func DownloadHandler(w http.ResponseWriter, req *http.Request, userID string) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	metadata, err := db.RetrieveVaultMetadata(id, userID)
	if err != nil {
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	}

	response := shared.VaultDownloadResponse{
		Name:         metadata.Name,
		ID:           metadata.ID,
		Chunks:       metadata.Chunks,
		Size:         metadata.Length,
		ProtectedKey: metadata.ProtectedKey,
	}

	jsonData, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

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

package send

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yeetfile/backend/cache"
	"yeetfile/backend/db"
	"yeetfile/backend/server/transfer"
	"yeetfile/backend/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

// UploadMetadataHandler handles a POST request to /u with the metadata required to set
// up a file for uploading. This is defined in the UploadMetadata struct.
func UploadMetadataHandler(w http.ResponseWriter, req *http.Request, _ string) {
	var meta shared.UploadMetadata
	data, _ := utils.LimitedReader(w, req.Body)
	err := json.Unmarshal(data, &meta)
	if err != nil {
		log.Printf("%v\n", req.Body)
		log.Printf("Error: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if meta.Chunks == 0 {
		http.Error(w, "# of chunks cannot be 0", http.StatusBadRequest)
		return
	} else if meta.Downloads == 0 {
		http.Error(w, "# of downloads cannot be 0", http.StatusBadRequest)
		return
	}

	_, err = UserCanSend(meta.Size, req)
	if err == OutOfSpaceError {
		http.Error(w, "Not enough space available", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	id, _ := db.InsertMetadata(meta.Chunks, meta.Name, meta.Salt, false)
	b2Upload := db.CreateNewUpload(id, meta.Name)

	exp := utils.StrToDuration(meta.Expiration)
	db.SetFileExpiry(id, meta.Downloads, time.Now().Add(exp).UTC())

	var b2Err error
	if meta.Chunks == 1 {
		b2Err = transfer.InitB2Upload(b2Upload)
	} else {
		b2Err = transfer.InitLargeB2Upload(meta.Name, b2Upload)
	}

	if b2Err != nil {
		http.Error(w, "Error initializing storage", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(shared.MetadataUploadResponse{ID: id})
	if err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
		return
	}
}

// UploadDataHandler handles the process of uploading file chunks to the server,
// after having already initialized the file metadata beforehand.
func UploadDataHandler(w http.ResponseWriter, req *http.Request, _ string) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-2]
	chunkNum, err := strconv.Atoi(segments[len(segments)-1])
	if err != nil {
		http.Error(w, "Invalid upload URL", http.StatusBadRequest)
		return
	}

	data, err := utils.LimitedReader(w, req.Body)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	metadata, err := db.RetrieveMetadata(id)
	if err != nil {
		http.Error(w, "No metadata found for file", http.StatusBadRequest)
		return
	}

	upload, b2Values, err := transfer.PrepareUpload(metadata, chunkNum, data)

	done, err := upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	// Update user meter
	err = UpdateUserMeter(len(data)-constants.TotalOverhead, req)
	if err != nil {
		// TODO: Maybe just silently accept this? Idk if it's worth an error
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	if done {
		_, _ = io.WriteString(w, id)
	}
}

// UploadPlaintextHandler handles uploading plaintext with a max size of
// shared.MaxPlaintextLen characters (shared/constants.go).
func UploadPlaintextHandler(w http.ResponseWriter, req *http.Request) {
	data, err := utils.LimitedReader(w, req.Body)
	var plaintextUpload shared.PlaintextUpload
	err = json.Unmarshal(data, &plaintextUpload)
	if err != nil {
		log.Printf("%v\n", req.Body)
		log.Printf("Error: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(plaintextUpload.Text) > constants.MaxPlaintextLen+constants.TotalOverhead {
		http.Error(w, "Invalid upload size", http.StatusBadRequest)
		return
	}

	id, _ := db.InsertMetadata(1, plaintextUpload.Name, plaintextUpload.Salt, true)
	b2Upload := db.CreateNewUpload(id, plaintextUpload.Name)

	exp := utils.StrToDuration(plaintextUpload.Expiration)
	db.SetFileExpiry(id, plaintextUpload.Downloads, time.Now().Add(exp).UTC())

	err = transfer.InitB2Upload(b2Upload)
	if err != nil {
		http.Error(w, "Unable to init file", http.StatusBadRequest)
		return
	}

	metadata, err := db.RetrieveMetadata(id)
	if err != nil {
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	}

	upload, b2Values, err := transfer.PrepareUpload(metadata, 1, plaintextUpload.Text)
	_, err = upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(shared.MetadataUploadResponse{ID: id})
	if err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
		return
	}
}

// DownloadHandler fetches metadata for downloading a file, such as the name of
// the file, the number of chunks, and the key for decrypting each chunk.
func DownloadHandler(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	metadata, err := db.RetrieveMetadata(id)
	if err != nil {
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	}

	expiry := db.GetFileExpiry(id)

	response := shared.DownloadResponse{
		Name:       metadata.Name,
		ID:         metadata.ID,
		Chunks:     metadata.Chunks,
		Salt:       metadata.Salt,
		Size:       metadata.Length,
		Downloads:  expiry.Downloads,
		Expiration: expiry.Date,
	}

	jsonData, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

// DownloadChunkHandler downloads individual chunks of a file using the chunk
// num from the file path and the decryption key in the header.
// Ex: /d/abc123/2 -- download the second chunk of file with id "abc123"
func DownloadChunkHandler(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")

	if len(segments) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := segments[len(segments)-2]
	chunk, _ := strconv.Atoi(segments[len(segments)-1])
	if chunk <= 0 {
		chunk = 1 // Downloads begin with chunk #1
	}

	metadata, err := db.RetrieveMetadata(id)
	if err != nil {
		http.Error(w, "No metadata found", http.StatusBadRequest)
		return
	}

	var eof bool
	var bytes []byte
	if cache.HasFile(id, metadata.Length) {
		eof, bytes = transfer.DownloadFileFromCache(id, metadata.Length, chunk)
	} else {
		cache.PrepCache(id, metadata.Length)
		eof, bytes = transfer.DownloadFile(metadata.B2ID, metadata.Length, chunk)
		_ = cache.Write(id, bytes)
	}

	// If the file is finished downloading, decrease the download counter
	// for that file, and delete if 0 are remaining
	rem := -1
	if eof {
		exp := db.GetFileExpiry(metadata.ID)
		rem = db.DecrementDownloads(metadata.ID)

		if rem == 0 {
			db.DeleteFileByMetadata(metadata)
		}

		if rem >= 0 {
			w.Header().Set("Downloads", strconv.Itoa(rem))
		}
		w.Header().Set("Date", fmt.Sprintf("%s", exp.Date.String()))
	}

	_, _ = w.Write(bytes)
}

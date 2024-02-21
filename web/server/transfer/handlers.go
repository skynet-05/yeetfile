package transfer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/utils"
)

// InitUploadHandler handles a POST request to /u with the metadata required to set
// up a file for uploading. This is defined in the UploadMetadata struct.
func InitUploadHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var meta shared.UploadMetadata
	err := decoder.Decode(&meta)
	if err != nil {
		log.Printf("%v\n", req.Body)
		log.Printf("Error: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	} else if !UserCanUpload(meta.Size, req) {
		http.Error(w, "Not enough space available", http.StatusBadRequest)
		return
	}

	id, _ := db.InsertMetadata(meta.Chunks, meta.Name, meta.Salt, false)
	b2Upload := db.CreateNewUpload(id)

	exp := utils.StrToDuration(meta.Expiration)
	db.SetFileExpiry(id, meta.Downloads, time.Now().Add(exp).UTC())

	if meta.Chunks == 1 {
		info, err := InitB2Upload()
		if err != nil {
			http.Error(w, "Unable to init file", http.StatusBadRequest)
			return
		}

		b2Upload.UpdateUploadValues(
			info.UploadURL,
			info.AuthorizationToken,
			info.BucketID)
	} else {
		info, err := InitLargeB2Upload(meta.Name)
		if err != nil {
			http.Error(w, "Unable to init file", http.StatusBadRequest)
			return
		}

		b2Upload.UpdateUploadValues(
			info.UploadURL,
			info.AuthorizationToken,
			info.FileID)
	}

	// Return ID to user
	_, _ = io.WriteString(w, id)
}

// UploadDataHandler handles the process of uploading file chunks to the server,
// after having already initialized the file metadata beforehand.
func UploadDataHandler(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-2]
	chunkNum, err := strconv.Atoi(segments[len(segments)-1])
	if err != nil {
		http.Error(w, "Invalid upload URL", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	upload, b2Values, err := PrepareUpload(id, chunkNum, data)
	done, err := upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	// Update user meter
	err = UpdateUserMeter(len(data)-shared.TotalOverhead, req)
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
	decoder := json.NewDecoder(req.Body)
	var plaintextUpload shared.PlaintextUpload
	err := decoder.Decode(&plaintextUpload)
	if err != nil {
		log.Printf("%v\n", req.Body)
		log.Printf("Error: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(plaintextUpload.Text) > shared.MaxPlaintextLen+shared.TotalOverhead {
		http.Error(w, "Invalid upload size", http.StatusBadRequest)
		return
	}

	id, _ := db.InsertMetadata(1, plaintextUpload.Name, plaintextUpload.Salt, true)
	b2Upload := db.CreateNewUpload(id)

	exp := utils.StrToDuration(plaintextUpload.Expiration)
	db.SetFileExpiry(id, plaintextUpload.Downloads, time.Now().Add(exp).UTC())

	info, err := InitB2Upload()
	if err != nil {
		http.Error(w, "Unable to init file", http.StatusBadRequest)
		return
	}

	b2Upload.UpdateUploadValues(
		info.UploadURL,
		info.AuthorizationToken,
		info.BucketID)

	upload, b2Values, err := PrepareUpload(id, 1, plaintextUpload.Text)
	_, err = upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	_, _ = io.WriteString(w, id)
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

	eof, bytes := DownloadFile(metadata.B2ID, metadata.Length, chunk)

	// If the file is finished downloading, decrease the download counter
	// for that file, and delete if 0 are remaining
	rem := -1
	if eof {
		exp := db.GetFileExpiry(metadata.ID)
		rem = db.DecrementDownloads(metadata.ID)

		if rem == 0 {
			db.DeleteFileByID(metadata.ID)
		}

		if rem >= 0 {
			w.Header().Set("Downloads", strconv.Itoa(rem))
		}
		w.Header().Set("Date", fmt.Sprintf("%s", exp.Date.String()))
	}

	_, _ = w.Write(bytes)
}

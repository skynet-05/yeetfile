package send

import (
	"log"
	"yeetfile/backend/db"
	"yeetfile/shared/constants"
)

func abortUpload(metadata db.FileMetadata, id string, dataLen, chunk int) {
	db.DeleteFileByMetadata(metadata)
	totalSize := dataLen
	for chunk > 1 {
		totalSize += constants.ChunkSize
		chunk--
	}

	err := UpdateUserMeter(-totalSize, id)
	if err != nil {
		log.Printf("Error updating user's meter during abort: %v\n", err)
	}
}

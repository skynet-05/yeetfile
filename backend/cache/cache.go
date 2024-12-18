package cache

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yeetfile/backend/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

var path = ".cache"
var enabled = true
var maxCacheSize int64 = 1024 * 1024 * 1024 * 25     // 25 gb max cache size
var maxCachedFileSize int64 = 1024 * 1024 * 1024 * 5 // 5 gb max file size

// Map file ID -> file size to ensure in-progress file caching is accounted for
// when determining if there's available space in the cache
var writeMap = map[string]int64{}
var accessMap = map[string]time.Time{}
var locks []string

func PrepCache(fileID string, size int64) {
	if !enabled || size > maxCachedFileSize || size > maxCacheSize || len(fileID) == 0 {
		return
	}

	totalCacheSize, err := utils.CheckDirSize(path)
	if err != nil {
		log.Printf(fmt.Sprintf("Unable to check cache dir size: %v", err))
		return
	}

	for totalCacheSize+size > int64(maxCacheSize) {
		err = removeOldestUnlockedFile()
		if err != nil {
			return
		}

		totalCacheSize, err = utils.CheckDirSize(path)
		if err != nil {
			return
		}
	}
	writeMap[fileID] = size
}

// HasFile returns true if the fileID provided exists in the cache and matches
// the expected size from the metadata table
func HasFile(fileID string, length int64) bool {
	if len(fileID) == 0 {
		return false
	}

	filePath := fmt.Sprintf("%s/%s", path, fileID)
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Ensure the file in the cache matches the size stored in the metadata
	// table
	return info.Size() == length
}

// Write writes file data to a cache file named with the file ID
func Write(fileID string, data []byte) error {
	if len(fileID) == 0 {
		return nil
	}

	_, found := writeMap[fileID]
	if !enabled || !found {
		return nil
	}

	filePath := fmt.Sprintf("%s/%s", path, fileID)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	if _, err = f.Write(data); err != nil {
		return err
	}

	return nil
}

// Read receives a file ID and start and end positions and reads from
// a file in the cache
func Read(fileID string, start int64, end int64) ([]byte, error) {
	if !enabled || len(fileID) == 0 {
		return nil, errors.New("cache not available")
	}

	filePath := fmt.Sprintf("%s/%s", path, fileID)
	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}

	// Ensure file ID is locked and cannot be deleted before reading is done
	if start == 0 {
		locks = append(locks, fileID)
	}

	// Update access time for this file on each read
	accessMap[fileID] = time.Now()

	var data []byte
	var err error
	if end < 0 {
		// Read full file
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		// Read part of file
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		data = make([]byte, end-start+1)
		_, err = file.ReadAt(data, start)
		if err != nil {
			return nil, err
		}
	}

	if start > 0 && end-start < int64(constants.ChunkSize) {
		// The last chunk of the file is being deleted, so the file lock
		// can be removed now
		locks = removeLock(locks, fileID)
	}

	return data, nil
}

// removeLock iterates through a list of locks and removes the first encountered
// instance of the locked file ID
func removeLock(currentLocks []string, id string) []string {
	for i, lock := range currentLocks {
		if lock == id {
			currentLocks[i] = currentLocks[len(currentLocks)-1]
			return currentLocks[:len(currentLocks)-1]
		}
	}

	return currentLocks
}

// removeOldestUnlockedFile iterates through the files in the cache and removes
// files with the oldest access times that are not currently locked.
func removeOldestUnlockedFile() error {
	if !enabled {
		return nil
	}

	var oldestFile string
	oldestTime := time.Now()
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories
			return nil
		}

		segments := strings.Split(filePath, "/")
		file := segments[len(segments)-1]

		if shared.ArrayContains(locks, file) {
			// Skip locked files
			return nil
		}

		fileTime, ok := accessMap[file]
		if !ok {
			fileTime = info.ModTime()
		}

		if fileTime.Before(oldestTime) {
			oldestTime = fileTime
			oldestFile = file
		}

		return err
	})

	if len(oldestFile) > 0 {
		err = os.Remove(fmt.Sprintf("%s/%s", path, oldestFile))
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveFile(id string) error {
	if !enabled || len(id) == 0 {
		return nil
	}

	filePath := fmt.Sprintf("%s/%s", path, id)
	if _, err := os.Stat(filePath); err != nil {
		// If the file doesn't exist, that's fine
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}

func init() {
	if os.Getenv("YEETFILE_STORAGE") == "local" {
		enabled = false
		return
	}

	userCacheDir := os.Getenv("YEETFILE_CACHE_DIR")
	if len(userCacheDir) > 0 {
		path = strings.TrimSuffix(userCacheDir, "/")
	}

	userCacheDirSize := os.Getenv("YEETFILE_CACHE_MAX_SIZE")
	if len(userCacheDirSize) > 0 {
		maxCacheSize = utils.ParseSizeString(userCacheDirSize)
		log.Printf("Max cache size: %s (%d bytes)",
			userCacheDirSize,
			maxCacheSize)
	} else {
		enabled = false
		return
	}

	userCacheFileSize := os.Getenv("YEETFILE_CACHE_MAX_FILE_SIZE")
	if len(userCacheFileSize) > 0 {
		maxCachedFileSize = utils.ParseSizeString(userCacheFileSize)
		log.Printf("Max size of files in cache: %s (%d bytes)",
			userCacheFileSize,
			maxCachedFileSize)
	} else {
		enabled = false
		return
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}

	log.Printf("Caching files to directory: %s", path)
}

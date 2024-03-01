package service

import (
	"github.com/benbusby/b2"
	"log"
	"os"
	"strconv"
	"strings"
	"yeetfile/web/utils"
)

var B2 b2.Auth
var B2BucketID string

const defaultStoragePath = "uploads"
const localStorageKey = "local"
const b2StorageKey = "b2"

func init() {
	B2BucketID = os.Getenv("B2_BUCKET_ID")

	if len(B2BucketID) == 0 {
		log.Fatal("Missing B2_BUCKET_ID environment variable")
	}

	var err error

	if strings.ToLower(os.Getenv("YEETFILE_STORAGE")) == localStorageKey {
		// Storage will bypass B2 and just store encrypted files on the
		// machine in the specified path or "uploads/"
		limit := 0
		limitStr := os.Getenv("YEETFILE_LOCAL_STORAGE_LIMIT")
		path := utils.GetEnvVar("YEETFILE_LOCAL_STORAGE_PATH", defaultStoragePath)

		if len(limitStr) > 0 {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				log.Fatalf("Invalid storage limit \"%s\"", limitStr)
			}
		}

		if limit > 0 {
			B2, err = b2.AuthorizeLimitedDummyAccount(path, limit)
		} else {
			B2, err = b2.AuthorizeDummyAccount(path)
		}
	} else {
		B2, err = b2.AuthorizeAccount(
			os.Getenv("B2_BUCKET_KEY_ID"),
			os.Getenv("B2_BUCKET_KEY"))
	}

	if err != nil {
		log.Fatalf("Unable to authenticate with B2: %v", err)
	}
}

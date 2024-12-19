package service

import (
	"github.com/benbusby/b2"
	"log"
	"os"
	"strconv"
	"yeetfile/backend/config"
	"yeetfile/backend/utils"
)

const defaultStoragePath = "uploads"

var (
	// public
	B2         *b2.Service
	B2BucketID string

	// private
	b2BucketKeyID string
	b2BucketKey   string
)

func init() {
	var err error

	if config.YeetFileConfig.StorageType == config.LocalStorage {
		log.Println("Setting up local storage...")
		// Storage will bypass B2 and just store encrypted files on the
		// machine in the specified path or "uploads/"
		var limit int64
		limitStr := os.Getenv("YEETFILE_LOCAL_STORAGE_LIMIT")
		path := utils.GetEnvVar("YEETFILE_LOCAL_STORAGE_PATH", defaultStoragePath)

		if len(limitStr) > 0 {
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				log.Fatalf("Invalid storage limit \"%s\"", limitStr)
			}
		}

		if limit > 0 {
			B2, err = b2.AuthorizeLimitedDummyAccount(path, limit)
		} else {
			B2, err = b2.AuthorizeDummyAccount(path)
		}
	} else if config.YeetFileConfig.StorageType == config.B2Storage {
		B2BucketID = utils.GetEnvVar("YEETFILE_B2_BUCKET_ID", "")
		b2BucketKeyID = utils.GetEnvVar("YEETFILE_B2_BUCKET_KEY_ID", "")
		b2BucketKey = utils.GetEnvVar("YEETFILE_B2_BUCKET_KEY", "")

		if len(B2BucketID) == 0 || len(b2BucketKeyID) == 0 || len(b2BucketKey) == 0 {
			log.Fatalf("Missing required B2 environment variables:\n"+
				"- YEETFILE_B2_BUCKET_ID: %v\n"+
				"- YEETFILE_B2_BUCKET_KEY_ID: %v\n"+
				"- YEETFILE_B2_BUCKET_KEY: %v\n",
				len(B2BucketID) > 0,
				len(b2BucketKeyID) > 0,
				len(b2BucketKey) > 0)
		}

		log.Println("Authorizing B2 account...")
		err = authorizeB2()
	} else {
		log.Fatalf("Invalid storage type '%s', "+
			"should be either '%s' or '%s'",
			config.YeetFileConfig.StorageType,
			config.B2Storage, config.LocalStorage)
	}

	if err != nil {
		log.Fatalf("Unable to authenticate with B2: %v", err)
	}
}

func AuthorizeB2() {
	_ = authorizeB2()
}

func authorizeB2() error {
	b2Tmp, _, err := b2.AuthorizeAccount(b2BucketKeyID, b2BucketKey)
	if err != nil {
		log.Println("Error authorizing B2 account", err)
		return err
	}

	// B2 authorized, replace global
	B2 = b2Tmp
	return nil
}

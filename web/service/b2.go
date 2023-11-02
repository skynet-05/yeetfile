package service

import (
	"github.com/benbusby/b2"
	"log"
	"os"
)

var B2 b2.Auth
var B2BucketID string

func init() {
	B2BucketID = os.Getenv("B2_BUCKET_ID")

	if len(B2BucketID) == 0 {
		log.Fatal("Missing B2_BUCKET_ID environment variable")
	}

	var err error
	B2, err = b2.AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))
	if err != nil {
		log.Fatalf("Unable to authenticate with B2: %v", err)
	}
}

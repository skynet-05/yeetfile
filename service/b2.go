package service

import (
	"os"
	"yeetfile/b2"
)

var B2 b2.Auth

func init() {
	var err error
	B2, err = b2.AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))
	if err != nil {
		panic(err)
	}
}

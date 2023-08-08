package main

import (
	"os"
	"yeetfile/db"
	"yeetfile/utils"
	"yeetfile/web/server"
)

var BucketID = os.Getenv("B2_BUCKET_ID")

func main() {
	go db.CheckExpiry()

	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port)
}

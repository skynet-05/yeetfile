package main

import (
	"os"
)

const domain string = "http://localhost:8090"
const ChunkSize int = 5242880

func main() {
	arg := os.Args[len(os.Args)-1]

	if arg == "config" {
		// TODO: Implement configurable backend
	} else if _, err := os.Stat(arg); err == nil {
		// Arg is a file that we should upload
		UploadFile(arg)
	} else {
		// Arg is (probably) a URL for a file
		StartDownload(arg)
	}
}

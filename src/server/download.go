package server

import (
	"fmt"
	"os"
	"yeetfile/src/backblaze"
	"yeetfile/src/utils"
)

func TestDownload() {
	auth, err := backblaze.B2AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))

	if err != nil {
		panic(err)
	}

	data, err := auth.B2DownloadById(os.Getenv("B2_TEST_UPLOAD_ID"))
	if err != nil {
		panic(err)
	}

	output, err := os.OpenFile("out.enc", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	_, _ = output.Write(data)

	out, err := os.ReadFile("out.enc")
	if err != nil {
		panic("Failed to read the output file")
	}

	plaintext := utils.Decrypt([]byte("topsecret"), out)

	fmt.Println(string(plaintext))
}

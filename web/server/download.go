package server

import (
	"golang.org/x/crypto/nacl/secretbox"
	"yeetfile/cli/crypto"
	"yeetfile/shared"
	"yeetfile/web/service"
)

type DownloadRequest struct {
	Password string `json:"password"`
}

func DownloadFile(
	b2ID string,
	length int,
	chunk int,
) (bool, []byte) {
	eof := false
	start := (chunk-1)*shared.ChunkSize +
		((crypto.NonceSize + secretbox.Overhead) * (chunk - 1))

	end := crypto.NonceSize +
		shared.ChunkSize +
		secretbox.Overhead +
		start - 1

	if end > length-1 {
		end = length - 1
		eof = true
	}

	data, _ := service.B2.PartialDownloadById(b2ID, start, end)

	return eof, data
}

package transfer

import (
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
		((shared.TotalOverhead) * (chunk - 1))

	end := shared.ChunkSize +
		shared.TotalOverhead +
		start - 1

	if end >= length-1 {
		end = length - 1
		eof = true
	}

	data, _ := service.B2.PartialDownloadById(b2ID, start, end)

	return eof, data
}

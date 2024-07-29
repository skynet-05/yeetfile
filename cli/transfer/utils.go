package transfer

import (
	"math"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

func GetNumChunks(size int64) int {
	return int(math.Ceil(float64(size) / float64(constants.ChunkSize)))
}

func GetReadBounds(chunk int, size int64) (int64, int64) {
	start := int64(constants.ChunkSize * chunk)
	end := int64(constants.ChunkSize * (chunk + 1))

	if end > size {
		end = size
	}

	return start, end
}

func GetModificationEndpoint(isFolder bool) endpoints.Endpoint {
	if isFolder {
		return endpoints.VaultFolder
	}

	return endpoints.VaultFile
}
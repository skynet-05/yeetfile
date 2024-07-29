package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/requests"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

type PendingDownload struct {
	ID                  string
	Key                 []byte
	File                *os.File
	NumChunks           int
	UnformattedEndpoint endpoints.Endpoint
	Server              string
}

type DownloadChunk struct {
	File     *os.File
	ChunkNum int
	Key      []byte
	Endpoint string
}

// worker sends chunked and encrypted file data to the endpoint specified in the
// provided FileChunk.
func downloadWorker(wCtx WorkerCtx, chunks <-chan DownloadChunk, progress func(), wg *sync.WaitGroup) {
	defer wg.Done()
	for chunk := range chunks {
		select {
		case <-wCtx.ctx.Done():
			log.Println("workers stopped due to cancellation")
			return
		default:
			data, err := fetchChunk(chunk)
			if err != nil {
				log.Printf("Worker error: %v\n", err)
				wCtx.cancel()
				return
			}

			err = writeChunk(chunk, data)
			progress()
		}
	}
}

func fetchChunk(chunk DownloadChunk) ([]byte, error) {
	resp, err := requests.GetRequest(chunk.Endpoint)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("server error")
	}

	body, _ := io.ReadAll(resp.Body)
	decryptedData, err := crypto.DecryptChunk(chunk.Key, body)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

func writeChunk(chunk DownloadChunk, data []byte) error {
	offset := int64(constants.ChunkSize * chunk.ChunkNum)
	_, err := chunk.File.WriteAt(data, offset)
	if err != nil {
		return err
	}

	return nil
}

func initDownload(
	id,
	server string,
	key []byte,
	file *os.File,
	chunks int,
) PendingDownload {
	return PendingDownload{
		ID:        id,
		Server:    server,
		Key:       key,
		File:      file,
		NumChunks: chunks,
	}
}

func InitSendDownload(
	id,
	server string,
	key []byte,
	file *os.File,
	chunks int,
) PendingDownload {
	p := initDownload(id, server, key, file, chunks)
	p.UnformattedEndpoint = endpoints.DownloadSendFileData
	return p
}

func InitVaultDownload(
	id string,
	key []byte,
	file *os.File,
) (PendingDownload, error) {
	url := endpoints.DownloadVaultFileMetadata.Format(config.UserConfig.Server, id)
	resp, err := requests.GetRequest(url)
	if err != nil {
		return PendingDownload{}, err
	} else if resp.StatusCode != http.StatusOK {
		return PendingDownload{}, utils.ParseHTTPError(resp)
	}

	var metadata shared.VaultDownloadResponse
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return PendingDownload{}, err
	}

	p := initDownload(id, config.UserConfig.Server, key, file, metadata.Chunks)
	p.UnformattedEndpoint = endpoints.DownloadVaultFileData
	return p, nil
}

func (p PendingDownload) DownloadData(progress func()) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wCtx := WorkerCtx{ctx: ctx, cancel: cancel}
	defer cancel()

	jobs := make(chan DownloadChunk, constants.MaxTransferThreads)
	for i := 1; i <= constants.MaxTransferThreads; i++ {
		wg.Add(1)
		go downloadWorker(wCtx, jobs, progress, &wg)
	}

	// Download all but the final file chunk using the workers
	for chunk := 0; chunk < p.NumChunks-1; chunk++ {
		chunkNum := strconv.Itoa(chunk + 1)
		url := p.UnformattedEndpoint.Format(p.Server, p.ID, chunkNum)
		fileChunk := DownloadChunk{
			File:     p.File,
			ChunkNum: chunk,
			Key:      p.Key,
			Endpoint: url,
		}
		jobs <- fileChunk
	}

	close(jobs)
	wg.Wait()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Download final chunk
	finalChunk := DownloadChunk{
		File:     p.File,
		ChunkNum: p.NumChunks - 1,
		Key:      p.Key,
		Endpoint: p.UnformattedEndpoint.Format(p.Server, p.ID, strconv.Itoa(p.NumChunks)),
	}
	data, err := fetchChunk(finalChunk)
	if err != nil {
		return err
	}

	err = writeChunk(finalChunk, data)
	if err != nil {
		return err
	}

	return nil
}

func DownloadText(id, server string, key []byte) ([]byte, error) {
	url := endpoints.DownloadSendFileData.Format(server, id, "1")
	resp, err := requests.GetRequest(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, utils.ParseHTTPError(resp)
	}

	body, _ := io.ReadAll(resp.Body)
	decryptedData, err := crypto.DecryptChunk(key, body)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

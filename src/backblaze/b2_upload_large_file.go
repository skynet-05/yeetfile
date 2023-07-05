package backblaze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

const APIStartLargeFile string = "b2_start_large_file"
const APIGetUploadPartURL string = "b2_get_upload_part_url"
const APIFinishLargeFile = "b2_finish_large_file"

type B2StartFile struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentSha1   string `json:"contentSha1"`
	ContentType   string `json:"contentType"`
	FileID        string `json:"fileId"`
	FileInfo      struct {
	} `json:"fileInfo"`
	FileName      string `json:"fileName"`
	FileRetention struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    struct {
			Mode                 any `json:"mode"`
			RetainUntilTimestamp any `json:"retainUntilTimestamp"`
		} `json:"value"`
	} `json:"fileRetention"`
	LegalHold struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    any  `json:"value"`
	} `json:"legalHold"`
	ServerSideEncryption struct {
		Algorithm any `json:"algorithm"`
		Mode      any `json:"mode"`
	} `json:"serverSideEncryption"`
	UploadTimestamp int64 `json:"uploadTimestamp"`
}

type B2FilePartInfo struct {
	FileID             string `json:"fileId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type B2LargeFile struct {
}

func (b2Auth B2Auth) B2StartLargeFile(
	filename string,
) (B2StartFile, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"bucketId": "%s",
		"fileName": "%s",
		"contentType": "b2/x-auto"
	}`, os.Getenv("B2_BUCKET_ID"), filename)))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, APIPrefix, APIStartLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return B2StartFile{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := B2Client.Do(req)
	if err != nil {
		log.Printf("Error starting B2 file: %v\n", err)
		return B2StartFile{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return B2StartFile{}, B2Error
	}

	var file B2StartFile
	err = json.NewDecoder(res.Body).Decode(&file)
	if err != nil {
		log.Printf("Error decoding B2 file init: %v", err)
		return B2StartFile{}, err
	}

	return file, nil
}

func (b2Auth B2Auth) B2GetUploadPartURL(
	b2File B2StartFile,
) (B2FilePartInfo, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s"
	}`, b2File.FileID)))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, APIPrefix, APIGetUploadPartURL)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return B2FilePartInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := B2Client.Do(req)
	if err != nil {
		log.Printf("Error getting B2 upload url: %v\n", err)
		return B2FilePartInfo{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return B2FilePartInfo{}, B2Error
	}

	var upload B2FilePartInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("Error decoding B2 upload part info: %v", err)
		return B2FilePartInfo{}, err
	}

	return upload, nil
}

func (b2PartInfo B2FilePartInfo) B2UploadFilePart(
	chunkNum int,
	checksum string,
	contents []byte,
) error {
	req, err := http.NewRequest(
		"POST",
		b2PartInfo.UploadURL,
		bytes.NewBuffer(contents))
	if err != nil {
		log.Printf("Error creating upload request: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization":     {b2PartInfo.AuthorizationToken},
		"Content-Length":    {strconv.Itoa(len(contents))},
		"X-Bz-Part-Number":  {strconv.Itoa(chunkNum)},
		"X-Bz-Content-Sha1": {checksum},
	}

	res, err := B2Client.Do(req)

	if err != nil {
		log.Printf("Error uploading file to B2: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", b2PartInfo.UploadURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return B2Error
	}

	return nil
}

func (b2Auth B2Auth) B2FinishLargeFile(
	fileID string,
	checksums string,
) error {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s",
		"partSha1Array": %s
	}`, fileID, checksums)))

	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, APIPrefix, APIFinishLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := B2Client.Do(req)

	if err != nil {
		log.Printf("Error finishing B2 upload: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return B2Error
	}

	return nil
}

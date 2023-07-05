package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

const AUTH_URL string = "https://api.backblazeb2.com/b2api/v2/b2_authorize_account"
const API_PREFIX string = "b2api/v2"
const API_START_LARGE_FILE string = "b2_start_large_file"
const API_GET_UPLOAD_PART_URL string = "b2_get_upload_part_url"
const API_UPLOAD_PART string = "b2_upload_part"
const API_FINISH_LARGE_FILE = "b2_finish_large_file"

var CLIENT = &http.Client{Timeout: 10 * time.Second}

type B2Auth struct {
	AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	AccountID               string `json:"accountId"`
	Allowed                 struct {
		BucketID     string   `json:"bucketId"`
		BucketName   string   `json:"bucketName"`
		Capabilities []string `json:"capabilities"`
		NamePrefix   any      `json:"namePrefix"`
	} `json:"allowed"`
	APIURL              string `json:"apiUrl"`
	AuthorizationToken  string `json:"authorizationToken"`
	DownloadURL         string `json:"downloadUrl"`
	RecommendedPartSize int    `json:"recommendedPartSize"`
	S3APIURL            string `json:"s3ApiUrl"`
}

type B2File struct {
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

type B2UploadInfo struct {
	FileID             string `json:"fileId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

func B2Init() (B2Auth, error) {
	req, err := http.NewRequest("GET", AUTH_URL, nil)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v", err)
		return B2Auth{}, err
	}

	authString := fmt.Sprintf(
		"%s:%s",
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))

	authString = base64.StdEncoding.EncodeToString([]byte(authString))

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic: %s", authString)},
	}

	res, err := CLIENT.Do(req)
	if err != nil {
		log.Printf("Error sending B2 auth request: %v", err)
		return B2Auth{}, err
	}

	var auth B2Auth
	err = json.NewDecoder(res.Body).Decode(&auth)
	if err != nil {
		log.Printf("Error decoding B2 auth: %v", err)
		return B2Auth{}, err
	}

	if strings.HasSuffix(auth.APIURL, "/") {
		auth.APIURL = auth.APIURL[0 : len(auth.APIURL)-2]
	}

	return auth, nil
}

func (b2Auth B2Auth) B2FileInit(filename string) (B2File, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"bucketId": "%s",
		"fileName": "%s",
		"contentType": "b2/x-auto"
	}`, os.Getenv("B2_BUCKET_ID"), filename)))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, API_PREFIX, API_START_LARGE_FILE)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return B2File{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := CLIENT.Do(req)
	if err != nil {
		log.Printf("Error sending B2 init file request: %v\n", err)
		return B2File{}, err
	}

	var file B2File
	err = json.NewDecoder(res.Body).Decode(&file)
	if err != nil {
		log.Printf("Error decoding B2 file init: %v", err)
		return B2File{}, err
	}

	return file, nil
}

func (b2Auth B2Auth) B2GetUploadURL(b2File B2File) (B2UploadInfo, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s"
	}`, b2File.FileID)))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, API_PREFIX, API_GET_UPLOAD_PART_URL)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return B2UploadInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := CLIENT.Do(req)
	if err != nil {
		log.Printf("Error sending B2 start upload request: %v\n", err)
		return B2UploadInfo{}, err
	}

	var upload B2UploadInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("Error decoding B2 upload init: %v", err)
		return B2UploadInfo{}, err
	}

	return upload, nil
}

func (b2Auth B2Auth) B2UploadFilePart(
	info B2UploadInfo,
	chunkNum int,
	checksum string,
	chunk []byte,
) error {
	//reqURL := fmt.Sprintf(
	//	"%s/%s/%s",
	//	b2Auth.APIURL, API_PREFIX, API_UPLOAD_PART)
	reqURL := info.UploadURL

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(chunk))
	if err != nil {
		log.Printf("Error creating request to upload chunk: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization":     {b2Auth.AuthorizationToken},
		"Content-Length":    {strconv.Itoa(len(chunk))},
		"X-Bz-Part-Number":  {strconv.Itoa(chunkNum)},
		"X-Bz-Content-Sha1": {checksum},
	}

	_, err = CLIENT.Do(req)
	if err != nil {
		log.Printf("Error uploading file chunk to B2: %v\n", err)
		return err
	}

	return nil
}

func (b2Auth B2Auth) B2FinishFileUpload(
	info B2UploadInfo,
	checksums string,
) error {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s",
		"partSha1Array": %s
	}`, info.FileID, checksums)))
	fmt.Println(reqBody)

	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, API_PREFIX, API_FINISH_LARGE_FILE)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := CLIENT.Do(req)

	resp, _ := httputil.DumpResponse(res, true)
	fmt.Println(fmt.Sprintf("%s", resp))
	if err != nil {
		log.Printf("Error sending B2 finish upload request: %v\n", err)
		return err
	}

	return nil
}

package backblaze

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

const AuthURL string = "https://api.backblazeb2.com/b2api/v2/b2_authorize_account"
const APIPrefix string = "b2api/v2"
const APIStartLargeFile string = "b2_start_large_file"
const APIGetUploadURL string = "b2_get_upload_url"
const APIGetUploadPartURL string = "b2_get_upload_part_url"
const APIFinishLargeFile = "b2_finish_large_file"

var HTTPClient = &http.Client{Timeout: 10 * time.Second}

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
	BucketID           string `json:"bucketId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type B2UploadPartInfo struct {
	FileID             string `json:"fileId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

func B2Init(b2BucketKeyId string, b2BucketKey string) (B2Auth, error) {
	req, err := http.NewRequest("GET", AuthURL, nil)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v", err)
		return B2Auth{}, err
	}

	authString := fmt.Sprintf("%s:%s", b2BucketKeyId, b2BucketKey)
	authString = base64.StdEncoding.EncodeToString([]byte(authString))

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic: %s", authString)},
	}

	res, err := HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error sending B2 auth request: %v", err)
		return B2Auth{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("%s -- error: %d\n", AuthURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
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

func (b2Auth B2Auth) B2StartLargeFile(filename string) (B2File, error) {
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
		return B2File{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error sending B2 init file request: %v\n", err)
		return B2File{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s -- error: %d\n", "POST", reqURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
	}

	var file B2File
	err = json.NewDecoder(res.Body).Decode(&file)
	if err != nil {
		log.Printf("Error decoding B2 file init: %v", err)
		return B2File{}, err
	}

	return file, nil
}

func (b2Auth B2Auth) B2GetUploadURL() (B2UploadInfo, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"bucketId": "%s"
	}`, os.Getenv("B2_BUCKET_ID"))))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, APIPrefix, APIGetUploadURL)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return B2UploadInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error requesting B2 upload URL: %v\n", err)
		return B2UploadInfo{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s -- error: %d\n", "POST", reqURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
	}

	var upload B2UploadInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("Error decoding B2 upload info: %v", err)
		return B2UploadInfo{}, err
	}

	return upload, nil
}

func (b2Auth B2Auth) B2GetUploadPartURL(b2File B2File) (B2UploadPartInfo, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s"
	}`, b2File.FileID)))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, APIPrefix, APIGetUploadPartURL)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return B2UploadPartInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error sending B2 start upload request: %v\n", err)
		return B2UploadPartInfo{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s -- error: %d\n", "POST", reqURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
	}

	var upload B2UploadPartInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("Error decoding B2 upload part info: %v", err)
		return B2UploadPartInfo{}, err
	}

	return upload, nil
}

func (b2Info B2UploadInfo) B2UploadFile(
	filename string,
	checksum string,
	data []byte,
) error {
	req, err := http.NewRequest("POST", b2Info.UploadURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request to upload chunk: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization":     {b2Info.AuthorizationToken},
		"Content-Type":      {"application/octet-stream"},
		"Content-Length":    {strconv.Itoa(len(data))},
		"X-Bz-File-Name":    {filename},
		"X-Bz-Content-Sha1": {checksum},
	}

	res, err := HTTPClient.Do(req)

	if err != nil {
		log.Printf("Error uploading file chunk to B2: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s -- error: %d\n", "POST", b2Info.UploadURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return errors.New("request returned error response")
	}

	return nil
}

func (b2Info B2UploadPartInfo) B2UploadFilePart(
	chunkNum int,
	checksum string,
	chunk []byte,
) error {
	req, err := http.NewRequest("POST", b2Info.UploadURL, bytes.NewBuffer(chunk))
	if err != nil {
		log.Printf("Error creating request to upload chunk: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization":     {b2Info.AuthorizationToken},
		"Content-Length":    {strconv.Itoa(len(chunk))},
		"X-Bz-Part-Number":  {strconv.Itoa(chunkNum)},
		"X-Bz-Content-Sha1": {checksum},
	}

	res, err := HTTPClient.Do(req)

	if err != nil {
		log.Printf("Error uploading file to B2: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s -- error: %d\n", "POST", b2Info.UploadURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
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

	res, err := HTTPClient.Do(req)

	if err != nil {
		log.Printf("Error sending B2 finish upload request: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s -- error: %d\n", "POST", reqURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
	}

	return nil
}

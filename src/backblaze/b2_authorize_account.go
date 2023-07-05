package backblaze

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

const AuthURL string = "https://api.backblazeb2.com/b2api/v2/b2_authorize_account"

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

func B2AuthorizeAccount(
	b2BucketKeyId string,
	b2BucketKey string,
) (B2Auth, error) {
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

	res, err := B2Client.Do(req)
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

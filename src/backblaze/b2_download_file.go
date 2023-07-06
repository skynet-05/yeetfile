package backblaze

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

const APIDownloadById string = "b2_download_file_by_id"

func (b2Auth B2Auth) B2DownloadById(id string) ([]byte, error) {
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, APIPrefix, APIDownloadById)

	req, err := http.NewRequest("GET", reqURL, nil)

	q := req.URL.Query()
	q.Add("fileId", id)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := B2Client.Do(req)
	if err != nil {
		log.Printf("Error requesting B2 download: %v\n", err)
		return nil, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "GET", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return nil, B2Error
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error reading response body")
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	return body, nil
}

package requests

import (
	"bytes"
	"net/http"
)

func GetRequest(session, url string) (*http.Response, error) {
	return sendRequest(session, http.MethodGet, url, nil)
}

func PostRequest(session, url string, data []byte) (*http.Response, error) {
	return sendRequest(session, http.MethodPost, url, data)
}

func PutRequest(session, url string, data []byte) (*http.Response, error) {
	return sendRequest(session, http.MethodPut, url, data)
}

func DeleteRequest(session, url string) (*http.Response, error) {
	return sendRequest(session, http.MethodDelete, url, nil)
}

func sendRequest(session, method, url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if err == nil && len(session) > 0 {
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: session,
		})
	}

	resp, err := new(http.Transport).RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

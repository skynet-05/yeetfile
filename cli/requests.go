package main

import (
	"bytes"
	"errors"
	"net/http"
)

var ServerError = errors.New("server error")

func GetRequest(url string) (*http.Response, error) {
	return sendRequest(http.MethodGet, url, nil)
}

func PostRequest(url string, data []byte) (*http.Response, error) {
	return sendRequest(http.MethodPost, url, data)
}

func PutRequest(url string, data []byte) (*http.Response, error) {
	return sendRequest(http.MethodPut, url, data)
}

func sendRequest(method string, url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if len(session) > 0 {
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

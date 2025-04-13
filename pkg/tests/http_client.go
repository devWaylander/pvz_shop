package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

const BaseURL = "http://localhost:8080"

type HttpClient struct {
	parent http.Client
}

func (client *HttpClient) SendJsonReq(token, method, url string, reqBody []byte) (resp *http.Response, resBody []byte, err error) { //nolint:stylecheck
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return client.sendReq(token, req)
}

func (client *HttpClient) sendReq(token string, req *http.Request) (resp *http.Response, resBody []byte, err error) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err = client.parent.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	resBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, resBody, nil
}

package request

import (
	"io"
	"net/http"
	"time"
)

var Headers map[string]string
var Timeout time.Duration

func POST(url string, body io.Reader) (*http.Response, error) {
	return newRequest("POST", url, body)
}

func GET(url string, body io.Reader) (*http.Response, error) {
	return newRequest("GET", url, body)
}
func newRequest(method string, url string, body io.Reader) (*http.Response, error) {
	Client := &http.Client{
		Timeout: Timeout,
	}
	req, err := http.NewRequest(method, url, body)
	for k, v := range Headers {
		req.Header.Set(k, v)
	}
	if err != nil {
		return nil, err
	}
	return Client.Do(req)
}

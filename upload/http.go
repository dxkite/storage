package upload

import (
	"errors"
	"io"
	"net/http"
)

var Client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}

func HttpGet(url string) (io.ReadCloser, error) {
	rr, er := Client.Get(url)
	if er != nil {
		return nil, er
	}
	if rr.StatusCode != http.StatusOK {
		return nil, errors.New("error http code")
	}
	return rr.Body, nil
}

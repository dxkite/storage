package common

import "net/http"

var Client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}

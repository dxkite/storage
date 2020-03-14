package main

import (
	dav "dxkite.cn/go-storage/src/webdav"
	"golang.org/x/net/webdav"
	"log"
	"net/http"
)

func main() {
	if err := http.ListenAndServe(":20214", &webdav.Handler{
		FileSystem: dav.FileSystem{Dir: "./data"},
		LockSystem: webdav.NewMemLS(),
	}); err != nil {
		log.Fatal(err)
	}
}

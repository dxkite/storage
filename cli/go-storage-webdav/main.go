package main

import (
	dav "dxkite.cn/go-storage/src/webdav"
	"golang.org/x/net/webdav"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	if err := http.ListenAndServe(":20214", &webdav.Handler{
		FileSystem: dav.NewSystem("./data"),
		LockSystem: webdav.NewMemLS(),
	}); err != nil {
		log.Fatal(err)
	}
}

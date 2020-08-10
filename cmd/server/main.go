package main

import (
	"dxkite.cn/storage"
	uploader "dxkite.cn/storage-uploader"
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", &storage.UploadHandler{
		Usn:       uploader.Default,
		BlockSize: 2 * 1024 * 1024,
	}))
}

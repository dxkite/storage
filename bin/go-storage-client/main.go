package main

import (
	"dxkite.cn/go-storage/lib/client"
	"log"
	"os"
)

func main() {
	f, _ := os.OpenFile("./data/example-2.gif", os.O_RDONLY, os.ModePerm)
	if er := client.UploadFile("127.0.0.1:8080", "example-2.gif", f); er != nil {
		log.Fatal("upload error:", er)
	}
}

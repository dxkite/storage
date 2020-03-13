package main

import (
	"dxkite.cn/go-storage/lib/client"
	"log"
	"os"
)

func main() {
	f, _ := os.OpenFile("./data/陈奕迅 - 十年.mp3", os.O_RDONLY, os.ModePerm)
	if er := client.UploadFile("127.0.0.1:8080", "陈奕迅 - 十年.mp3", f); er != nil {
		log.Fatal("upload error:", er)
	}
}

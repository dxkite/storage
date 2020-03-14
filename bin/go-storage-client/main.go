package main

import (
	"dxkite.cn/go-storage/src/client"
	"encoding/hex"
	"log"
	"time"
)

func Upload(path string) {
	u := client.NewUploader("127.0.0.1:8080", time.Second*100)
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
}

func Download(info, path string) {
	h, _ := hex.DecodeString(info)
	d := client.NewDownloader("127.0.0.1:8080", h)
	if er := d.DownloadToFile(path); er != nil {
		log.Fatal(er)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	Upload("./data/陈奕迅 - 十年.mp3")
	Download("0d14a3da07c74efaee62f3ea495ce7de2e62c257", "data/download.mp3")
}

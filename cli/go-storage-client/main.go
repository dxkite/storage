package main

import (
	"dxkite.cn/go-storage/src/client"
	"encoding/hex"
	"flag"
	"log"
	"time"
)

func Upload(addr, path string) {
	u := client.NewUploader(addr, time.Second*100)
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
	log.Println("upload success")
}

func Download(addr, info, path string) {
	h, _ := hex.DecodeString(info)
	d := client.NewDownloader(addr, h)
	if er := d.DownloadToFile(path); er != nil {
		log.Fatal(er)
	}
	log.Println("download success")
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:20214", "listening")
	var path = flag.String("path", "./", "download to path")
	var help = flag.Bool("help", false, "print help info")

	flag.Parse()
	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}

	name := flag.Arg(0)

	if client.FileExist(name) {
		Upload(*addr, name)
	} else {
		Download(*addr, name, *path)
	}
}

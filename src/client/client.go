package client

import (
	"dxkite.cn/go-storage/src/install"
	"log"
)

type Client struct {
}

func (Client) Upload(path string, bs int) {
	u := NewUploader(int64(bs*1024*1024), "ali")
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
	log.Println("upload success")
}

func Install(path string) {
	if er := install.CreateHelper(path); er != nil {
		log.Println("error install", er)
	}
}

func (Client) Download(meta, path string, check bool, num, retry int) {
	d := NewMetaDownloader(meta, check, num)
	for retry > 0 {
		if er := d.DownloadToFile(path); er != nil {
			retry--
			log.Println("download error", er, "retry", retry)
		} else {
			break
		}
	}
	log.Println("download success")
}

var Default = Client{}

package client

import "log"

type Client struct {
}

func (Client) Upload(path string, bs int) {
	u := NewUploader(int64(bs*1024*1024), "ali")
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
	log.Println("upload success")
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

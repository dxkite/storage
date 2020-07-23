package client

import (
	"dxkite.cn/storage/client/downloader"
	"dxkite.cn/storage/client/uploader"
	"dxkite.cn/storage/util"
	"log"
)

type Client struct {
}

func (Client) Upload(path, cloud string, bs int) {
	u := uploader.NewUploader(int64(bs*1024*1024), cloud)
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
	log.Println("upload success")
}

func Install(path string) {
	if er := util.Install(path); er != nil {
		log.Println("error install", er)
	} else {
		log.Println("install success")
	}
}

func Uninstall(path string) {
	if er := util.Uninstall(path); er != nil {
		log.Println("error uninstall", er)
	} else {
		log.Println("uninstall success")
	}
}

func (Client) Download(meta, path string, check bool, num, retry int) {
	d := downloader.NewMetaDownloader(check, num)

	if err := d.Load(meta); err != nil {
		log.Println("load meta error", err)
		return
	}

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

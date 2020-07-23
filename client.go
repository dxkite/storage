package storage

import (
	"log"
)

func Upload(path, cloud string, bs int) {
	u := NewUploader(int64(bs*1024*1024), cloud)
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
	log.Println("upload success")
}

func Install(path string) {
	if er := InstallURL(path); er != nil {
		log.Println("error install", er)
	} else {
		log.Println("install success")
	}
}

func Uninstall(path string) {
	if er := UninstallURL(path); er != nil {
		log.Println("error uninstall", er)
	} else {
		log.Println("uninstall success")
	}
}

func Download(meta, path string, check bool, num, retry int) {
	d := NewMetaDownloader(check, num)

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

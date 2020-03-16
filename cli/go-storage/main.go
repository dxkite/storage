package main

import (
	"dxkite.cn/go-storage/src/client"
	"flag"
	"log"
)

func Upload(path string) {
	u := client.NewUploader(2*1024*1024, "ali")
	if er := u.UploadFile(path); er != nil {
		log.Fatal("upload error:", er)
	}
	log.Println("upload success")
}

func DownloadMeta(meta, path string) {
	d := client.NewMetaDownloader(meta)
	if er := d.DownloadToFile(path); er != nil {
		log.Fatal(er)
	}
	log.Println("download success")
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var path = flag.String("path", "./", "download to path")
	var meta = flag.Bool("meta", false, "use meta file")
	var help = flag.Bool("help", false, "print help info")

	flag.Parse()
	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}

	name := flag.Arg(0)

	if client.FileExist(name) {
		if *meta {
			DownloadMeta(name, *path)
		} else {
			Upload(name)
		}
	} else {
		log.Println("error input file", name)
	}
}

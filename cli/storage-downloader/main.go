package main

import (
	"dxkite.cn/go-storage/src/client"
	"flag"
	"log"
)

func DownloadMeta(meta, path string, check bool) {
	d := client.NewMetaDownloader(meta, check)
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
	var help = flag.Bool("help", false, "print help info")
	var check = flag.Bool("check", false, "check hash after download")

	flag.Parse()
	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}
	name := flag.Arg(0)
	DownloadMeta(name, *path, *check)
}

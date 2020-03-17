package main

import (
	"dxkite.cn/go-storage/src/client"
	"flag"
	"log"
)

func DownloadMeta(meta, path string, check bool, num int) {
	d := client.NewMetaDownloader(meta, check, num)
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
	var num = flag.Int("threads", 20, "max download threads")

	flag.Parse()
	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}
	name := flag.Arg(0)
	DownloadMeta(name, *path, *check, *num)
}

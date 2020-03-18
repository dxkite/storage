package main

import (
	"dxkite.cn/go-storage/src/client"
	"flag"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var path = flag.String("path", "./", "download to path")
	var meta = flag.Bool("meta", false, "use meta file")
	var uncheck = flag.Bool("uncheck", false, "uncheck hash after downloaded")
	var block = flag.Int("block_size", 2, "block size, mb")
	var help = flag.Bool("help", false, "print help info")
	var num = flag.Int("threads", 20, "max download threads")
	var retry = flag.Int("retry", 20, "max retry when error")

	flag.Parse()
	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}

	name := flag.Arg(0)
	if client.FileExist(name) {
		if *meta {
			client.Default.Download(name, *path, *uncheck == false, *num, *retry)
		} else {
			client.Default.Upload(name, *block)
		}
	} else {
		client.Default.Download(name, *path, *uncheck == false, *num, *retry)
	}
}

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
	var help = flag.Bool("help", false, "print help info")
	var uncheck = flag.Bool("uncheck", false, "uncheck hash after downloaded")
	var num = flag.Int("threads", 20, "max download threads")
	var retry = flag.Int("retry", 20, "max retry when error")

	flag.Parse()
	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}
	name := flag.Arg(0)
	client.Default.Download(name, *path, *uncheck == false, *num, *retry)
}

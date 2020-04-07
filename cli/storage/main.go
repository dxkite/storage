package main

import (
	"dxkite.cn/go-storage/src/client"
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var dl = flag.String("dl", "", "download to dl")
	var meta = flag.Bool("meta", false, "use meta file")
	var install = flag.Bool("install", false, "install")
	var uncheck = flag.Bool("uncheck", false, "uncheck hash after downloaded")
	var block = flag.Int("block_size", 2, "block size, mb")
	var help = flag.Bool("help", false, "print help info")
	var num = flag.Int("threads", 20, "max download threads")
	var retry = flag.Int("retry", 20, "max retry when error")

	flag.Parse()
	p, _ := filepath.Abs(os.Args[0])

	if *install {
		client.Install(p)
		return
	}

	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}

	name := flag.Arg(0)
	if client.FileExist(name) {
		if *meta {
			p, _ := filepath.Abs(name)
			*dl = filepath.Dir(p)
			client.Default.Download(name, *dl, *uncheck == false, *num, *retry)
		} else {
			client.Default.Upload(name, *block)
		}
	} else {
		if len(*dl) <= 0 {
			pp := filepath.Dir(p)
			*dl = path.Join(pp, "Download")
			_ = os.MkdirAll(*dl, os.ModePerm)
		}
		client.Default.Download(name, *dl, *uncheck == false, *num, *retry)
	}
}

package main

import (
	"dxkite.cn/go-storage/src/client"
	"dxkite.cn/go-storage/src/common"
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var save = flag.String("path", "", "download save path")

	var install = flag.Bool("install", false, "install")
	var uninstall = flag.Bool("uninstall", false, "uninstall")

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

	if *uninstall {
		client.Uninstall(p)
		return
	}

	if *help || flag.NArg() < 1 {
		flag.Usage()
		return
	}

	name := flag.Arg(0)
	if common.FileExist(name) {
		if strings.HasSuffix(name, common.EXT_META) {
			if len(*save) == 0 {
				p, _ := filepath.Abs(name)
				*save = filepath.Dir(p)
			}
			client.Default.Download(name, *save, *uncheck == false, *num, *retry)
		} else {
			client.Default.Upload(name, *block)
		}
	} else {
		if len(*save) == 0 {
			pp := filepath.Dir(p)
			*save = path.Join(pp, "Download")
			_ = os.MkdirAll(*save, os.ModePerm)
		}
		client.Default.Download(name, *save, *uncheck == false, *num, *retry)
	}
}

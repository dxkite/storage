package main

import (
	"dxkite.cn/storage"
	uploader "dxkite.cn/storage-uploader"
	"flag"
	"log"
	"net/http"
	"path"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var usn = flag.String("usn", uploader.Default, "upload usn")
	var auth = flag.String("auth", "", "auth api")
	var field = flag.String("auth_field", "token", "auth api field")
	var root = flag.String("root", "data", "upload root path")
	var block = flag.Int("block_size", 2, "block size, mb")
	var addr = flag.String("addr", ":8080", "listen addr")
	var help = flag.Bool("help", false, "print help info")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	log.Println("start server at", *addr)
	if len(*auth) > 0 {
		log.Println("enable auth", *auth)
	}

	log.Fatal(http.ListenAndServe(*addr, &storage.UploadHandler{
		Usn:        *usn,
		BlockSize:  *block * 1024 * 1024,
		AuthRemote: *auth,
		AuthField:  *field,
		Temp:       path.Join(*root, ".tmp"),
		Root:       path.Join(*root, "storage"),
	}))
}

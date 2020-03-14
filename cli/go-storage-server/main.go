package main

import (
	"dxkite.cn/go-storage/src/server"
	"dxkite.cn/go-storage/storage"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	var blockSize = flag.Int("size", 2*1024*1024, "block size")
	var path = flag.String("data", "./data", "data path")
	var addr = flag.String("listen", ":20214", "listening")
	var help = flag.Bool("help", false, "print help info")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	s := server.New(*path, int64(*blockSize))
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("go-storage", "v"+s.Version, "start, listen at", lis.Addr())
	gs := grpc.NewServer()
	storage.RegisterGoStorageServer(gs, s)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

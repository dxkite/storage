package main

import (
	"dxkite.cn/go-storage/src/server"
	"dxkite.cn/go-storage/storage"
	"google.golang.org/grpc"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	s := server.New("./data", 2*1024*1024)
	lis, err := net.Listen("tcp", ":8080")
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

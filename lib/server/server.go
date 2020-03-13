package server

import (
	"context"
	"dxkite.cn/go-storage/storage"
)

type GoStorageServer struct {
	storage.UnimplementedGoStorageServer

	Version   string // 版本
	BlockSize int64  // 块大小
	FreeSize  int64  // 空闲空间大小
	Path      string // 元数据存储位置
}

func (s *GoStorageServer) Hello(ctx context.Context, req *storage.PingRequest) (*storage.PongResponse, error) {
	return &storage.PongResponse{
		Version: s.Version,
		Size:    s.FreeSize,
		Block:   s.BlockSize,
	}, nil
}

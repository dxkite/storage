package server

import (
	"context"
	"dxkite.cn/go-storage/lib/encoding"
	"dxkite.cn/go-storage/lib/meta"
	"dxkite.cn/go-storage/storage"
	"dxkite.cn/go-storage/upload"
	"encoding/hex"
	"errors"
	"os"
	"path"
)

type GoStorageServer struct {
	storage.UnimplementedGoStorageServer

	Version   string // 版本
	BlockSize int64  // 块大小
	FreeSize  int64  // 空闲空间大小
	Root      string // 元数据存储位置
}

func New(root string, size int64) *GoStorageServer {
	if e := os.MkdirAll(root, os.ModePerm); e != nil {
		panic(e)
	}
	return &GoStorageServer{
		Version:   "1.0",
		BlockSize: size,
		FreeSize:  0,
		Root:      root,
	}
}

func (s *GoStorageServer) Hello(ctx context.Context, req *storage.PingRequest) (*storage.PongResponse, error) {
	return &storage.PongResponse{
		Version: s.Version,
		Size:    s.FreeSize,
		Block:   s.BlockSize,
	}, nil
}

func (s *GoStorageServer) Create(ctx context.Context, req *storage.StorageCreateRequest) (*storage.StorageResponse, error) {
	if len(req.Info) != 20 {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_HASH,
			Message: "error hash, need sha1 len 20",
		}, nil
	}

	m := meta.MetaInfo{
		Hash:   req.Info,
		Name:   req.Name,
		Status: meta.Create,
		Size:   req.Size,
		Type:   int32(storage.DataResponse_URI),
		Encode: int32(storage.DataResponse_IMAGE),
		Block:  nil,
	}

	f := path.Join(s.Root, hex.Dump(req.Info)+".meta")
	e := meta.EncodeToFile(f, &m)
	if e != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_STORE,
			Message: e.Error(),
		}, nil
	}

	return &storage.StorageResponse{
		Code:    storage.StorageResponse_ERROR_NONE,
		Message: "create success",
	}, nil
}

func (s *GoStorageServer) Store(ctx context.Context, req *storage.StorageStoreRequest) (*storage.StorageResponse, error) {

	if len(req.Hash) != 20 {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_HASH,
			Message: "error hash, need sha1 len 20",
		}, nil
	}

	if len(req.Info) != 20 {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_HASH,
			Message: "error hash, need sha1 len 20",
		}, nil
	}

	f := path.Join(s.Root, hex.Dump(req.Info)+".meta")
	m, e := meta.DecodeToFile(f)
	if e != nil && e == os.ErrNotExist {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_HASH,
			Message: "unknown hash",
		}, nil
	}

	if e != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_UNKNOWN,
			Message: e.Error(),
		}, nil
	}

	// 编码成图片
	b, eer := encoding.EncodeByte(req.Data)
	if eer != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_UNKNOWN,
			Message: eer.Error(),
		}, nil
	}

	// 上传到阿里云
	rt, uer := upload.Upload("ali", &upload.FileObject{
		Name: hex.Dump(req.Hash) + ".png",
		Data: b,
	})

	if uer != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_STORE,
			Message: uer.Error(),
		}, nil
	}

	m.Status = meta.Uploading
	m.AppendBlock(&meta.DataBlock{
		Hash:  req.Hash,
		Index: req.Index,
		Data:  []byte(rt.Url),
	})

	e = meta.EncodeToFile(f, m)
	if e != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_STORE,
			Message: e.Error(),
		}, nil
	}

	return &storage.StorageResponse{
		Code:    storage.StorageResponse_ERROR_NONE,
		Message: "block upload success",
	}, nil
}

func (s *GoStorageServer) Finish(ctx context.Context, req *storage.StorageFinishRequest) (*storage.StorageResponse, error) {
	if len(req.Info) != 20 {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_HASH,
			Message: "error hash, need sha1 len 20",
		}, nil
	}

	f := path.Join(s.Root, hex.Dump(req.Info)+".meta")
	m, e := meta.DecodeToFile(f)
	if e != nil && e == os.ErrNotExist {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_HASH,
			Message: "unknown hash",
		}, nil
	}

	if e != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_UNKNOWN,
			Message: e.Error(),
		}, nil
	}

	m.Status = meta.Finish

	e = meta.EncodeToFile(f, m)
	if e != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_STORE,
			Message: e.Error(),
		}, nil
	}

	return &storage.StorageResponse{
		Code:    storage.StorageResponse_ERROR_NONE,
		Message: "finish upload",
	}, nil
}

func (s *GoStorageServer) Get(ctx context.Context, req *storage.GetResponse) (*storage.DataResponse, error) {
	if len(req.Info) != 20 {
		return nil, nil
	}
	f := path.Join(s.Root, hex.Dump(req.Info)+".meta")
	m, e := meta.DecodeToFile(f)
	if e != nil && e == os.ErrNotExist {
		return nil, errors.New("file " + hex.Dump(req.Info) + " not found")
	}
	if e != nil {
		return nil, e
	}
	return NewDataResponse(m), nil
}

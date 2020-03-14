package server

import (
	"context"
	"dxkite.cn/go-storage/src/image"
	"dxkite.cn/go-storage/src/meta"
	"dxkite.cn/go-storage/src/upload"
	"dxkite.cn/go-storage/storage"
	"encoding/hex"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
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
		return nil, errors.New("error hash info")
	}

	m := meta.MetaInfo{
		Hash:      req.Info,
		Name:      req.Name,
		Status:    meta.Create,
		Size:      req.Size,
		Type:      int32(storage.DataResponse_URI),
		Encode:    int32(storage.DataResponse_IMAGE),
		Block:     nil,
		BlockSize: s.BlockSize,
	}

	f := path.Join(s.Root, fmt.Sprintf("%x.meta", req.Info))
	log.Printf("create meta %s\n", f)

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
		return nil, errors.New("error hash info")
	}

	if len(req.Info) != 20 {
		return nil, errors.New("error hash info")
	}

	f := path.Join(s.Root, fmt.Sprintf("%x.meta", req.Info))
	m, e := meta.DecodeFromFile(f)
	if e != nil && e == os.ErrNotExist {
		return nil, status.Errorf(codes.NotFound, "file "+fmt.Sprintf("%x", req.Info)+" not found")
	}

	if e != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_UNKNOWN,
			Message: e.Error(),
		}, nil
	}
	log.Printf("store meta %s index %d\n", f, req.Index)
	// 编码成图片
	b, eer := image.EncodeByte(req.Data)
	if eer != nil {
		return &storage.StorageResponse{
			Code:    storage.StorageResponse_ERROR_UNKNOWN,
			Message: eer.Error(),
		}, nil
	}
	log.Println("uploading")
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
	log.Println("uploaded to cloud")
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
	log.Println("restore meta")
	return &storage.StorageResponse{
		Code:    storage.StorageResponse_ERROR_NONE,
		Message: "block upload success",
	}, nil
}

func (s *GoStorageServer) Finish(ctx context.Context, req *storage.StorageFinishRequest) (*storage.StorageResponse, error) {
	if len(req.Info) != 20 {
		return nil, errors.New("error hash info")
	}
	f := path.Join(s.Root, fmt.Sprintf("%x.meta", req.Info))
	m, e := meta.DecodeFromFile(f)
	if e != nil && e == os.ErrNotExist {
		return nil, status.Errorf(codes.NotFound, "file "+fmt.Sprintf("%x", req.Info)+" not found")
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
	log.Printf("finish meta %s\n", f)
	return &storage.StorageResponse{
		Code:    storage.StorageResponse_ERROR_NONE,
		Message: "finish upload",
	}, nil
}

func (s *GoStorageServer) Get(ctx context.Context, req *storage.GetResponse) (*storage.DataResponse, error) {
	if len(req.Info) != 20 {
		return nil, errors.New("error hash info")
	}
	f := path.Join(s.Root, fmt.Sprintf("%x.meta", req.Info))
	m, e := meta.DecodeFromFile(f)
	if e != nil && e == os.ErrNotExist {
		return nil, status.Errorf(codes.NotFound, "file "+fmt.Sprintf("%x", req.Info)+" not found")
	}
	if e != nil {
		return nil, e
	}
	if m.Status != meta.Finish {
		return nil, status.Errorf(codes.NotFound, "file "+fmt.Sprintf("%d:%x", m.Status, req.Info)+" not found")
	}
	log.Printf("get meta %x\n", req.Info)
	return NewDataResponse(m), nil
}

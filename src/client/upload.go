package client

import (
	"context"
	"dxkite.cn/go-storage/storage"
	"errors"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

type Uploader struct {
	Info    []byte
	Path    string
	Size    int64
	Remote  string
	Timeout time.Duration

	c   storage.GoStorageClient
	ctx context.Context
}

func NewUploader(remote string, time time.Duration) *Uploader {
	return &Uploader{
		Remote:  remote,
		Timeout: time,
	}
}

func (u *Uploader) UploadFile(name string) error {
	file, oer := os.OpenFile(name, os.O_RDONLY, os.ModePerm)
	if oer != nil {
		return oer
	}

	conn, err := grpc.Dial(u.Remote, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()
	c := storage.NewGoStorageClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), u.Timeout)
	defer cancel()
	r, er := c.Hello(ctx, &storage.PingRequest{})
	if er != nil {
		return er
	}
	var info = SteamHash(file)
	var size = SteamSize(file)

	u.Info = info
	u.Size = size
	u.c = c
	u.ctx = ctx

	if e := u.SendCreate(file.Name(), info, size); e != nil {
		return e
	}
	log.Println("created")
	var buf = make([]byte, r.Block)
	var index = int64(0)
	for {
		nr, er := file.Read(buf)
		if nr > 0 {
			if err = u.SendStore(index, info, buf); err == nil {
				log.Printf("upload %d block success\n", index)
				index += 1
			} else {
				return err
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	if err != nil {
		return err
	}
	if e := u.SendFinish(info); e != nil {
		return e
	}
	log.Println("finished")
	return nil
}

func (u *Uploader) SendCreate(name string, info []byte, size int64) error {
	sr, er := u.c.Create(u.ctx, &storage.StorageCreateRequest{
		Info: info,
		Size: size,
		Name: name,
	})
	if er != nil {
		return er
	}
	if sr.Code != storage.StorageResponse_ERROR_NONE {
		return errors.New(sr.Message)
	}
	return nil
}

func (u *Uploader) SendStore(index int64, info, data []byte) error {
	sr, er := u.c.Store(u.ctx, &storage.StorageStoreRequest{
		Info:  info,
		Hash:  ByteHash(data),
		Data:  data,
		Index: index,
	})
	if er != nil {
		return er
	}
	if sr.Code != storage.StorageResponse_ERROR_NONE {
		return errors.New(sr.Message)
	}
	return nil
}

func (u *Uploader) SendFinish(info []byte) error {
	sr, er := u.c.Finish(u.ctx, &storage.StorageFinishRequest{
		Info: info,
	})
	if er != nil {
		return er
	}
	if sr.Code != storage.StorageResponse_ERROR_NONE {
		return errors.New(sr.Message)
	}
	return nil
}

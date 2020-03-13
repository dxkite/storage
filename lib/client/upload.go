package client

import (
	"context"
	"crypto/sha1"
	"dxkite.cn/go-storage/storage"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func UploadFile(remote, name string, file io.ReadSeeker) error {
	conn, err := grpc.Dial(remote, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()
	c := storage.NewGoStorageClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	r, er := c.Hello(ctx, &storage.PingRequest{})
	if er != nil {
		return er
	}
	var info = SteamHash(file)
	var size = SteamSize(file)
	if e := SendCreate(c, ctx, name, info, size); e != nil {
		return e
	}
	log.Println("created")
	var buf = make([]byte, r.Block)
	var index = int64(0)
	for {
		nr, er := file.Read(buf)
		if nr > 0 {
			if err = SendStore(c, ctx, index, info, buf); err == nil {
				log.Printf("upload %d piece success\n", index)
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
	if e := SendFinish(c, ctx, info); e != nil {
		return e
	}
	log.Println("finished")
	return nil
}

func SteamHash(r io.ReadSeeker) []byte {
	h := sha1.New()
	_, _ = r.Seek(0, io.SeekStart)
	_, err := io.Copy(h, r)
	if err != nil {
		panic(fmt.Sprintf("check sum: %v", err))
	}
	_, _ = r.Seek(0, io.SeekStart)
	return h.Sum(nil)
}

func ByteHash(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	return h.Sum(nil)
}

func SteamSize(r io.ReadSeeker) int64 {
	n, _ := r.Seek(0, io.SeekEnd)
	_, _ = r.Seek(0, io.SeekStart)
	return n
}

func SendCreate(c storage.GoStorageClient, ctx context.Context, name string, info []byte, size int64) error {
	sr, er := c.Create(ctx, &storage.StorageCreateRequest{
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

func SendStore(c storage.GoStorageClient, ctx context.Context, index int64, info, data []byte) error {
	sr, er := c.Store(ctx, &storage.StorageStoreRequest{
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

func SendFinish(c storage.GoStorageClient, ctx context.Context, info []byte) error {
	sr, er := c.Finish(ctx, &storage.StorageFinishRequest{
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

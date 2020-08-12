package storage

import (
	"crypto/sha1"
	"dxkite.cn/storage/image"
	"dxkite.cn/storage/meta"
	"dxkite.cn/storage/upload"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var DefaultUSN = os.Getenv(ENV_USN)

type Uploader struct {
	Size      int64
	Usn       string
	Notify    UploadNotify
	Processor UploadProcessor
}

func NewUploader(bs int64, usn string) *Uploader {
	return &Uploader{
		Size: bs,
		Usn:  usn,
	}
}

func (u *Uploader) UploadFile(p string) error {
	file, oer := os.OpenFile(p, os.O_RDONLY, os.ModePerm)
	if oer != nil {
		return oer
	}
	size := SteamSize(file)
	name := file.Name()
	base := filepath.Base(name)
	u.Notify = &ConsoleNotify{}
	u.Processor = NewFileUploadProcessor(name+EXT_UPLOADING, NewUploadInfo(base, size, u.Size))
	if mt, err := u.UploadStream(file); err != nil {
		return err
	} else {
		if er := meta.EncodeToFile(name+EXT_META, mt); er != nil {
			return er
		}
	}
	return nil
}

func (u *Uploader) UploadStream(r io.Reader) (*meta.Info, error) {
	log.Printf("upload to %s\n", u.Usn)
	p := u.Notify
	s := u.Processor

	ui, _ := s.Load()

	var index = int64(0)
	var err error
	var buf = make([]byte, u.Size)
	var uploader upload.Uploader

	if u, er := upload.Create(u.Usn); er != nil {
		return nil, er
	} else {
		uploader = u
	}

	sh := sha1.New()
	for {
		start, end := IndexToRange(ui.Meta.Size, ui.Meta.BlockSize, index)
		if start > ui.Meta.Size {
			break
		}
		n := int(end - start)
		nr, er := io.ReadAtLeast(r, buf, n)
		if nr > 0 {
			sh.Write(buf[:nr])
			if ui.Index.Get(index) {
				_ = p.Process(PROCESS_EXIST, index, start, end, nil)
				index++
				continue
			}
			_ = p.Process(PROCESS_START, index, start, end, nil)
			hh := ByteHash(buf[:nr])
			var encoded []byte
			if b, er := image.EncodeByte(buf[:nr]); er != nil {
				err = er
				break
			} else {
				encoded = b
			}
			if r, er := uploader.Upload(&upload.FileObject{
				Name: fmt.Sprintf("%s-%d.jpg", ui.Meta.Name, index),
				Data: encoded,
			}); er != nil {
				err = er
				_ = p.Process(PROCESS_ERROR, index, start, end, err)
				break
			} else {
				ui.Meta.Status = meta.Uploading
				b := meta.DataBlock{
					Hash:  hh,
					Index: index,
					Data:  []byte(r.Url),
				}
				ui.Meta.Block = append(ui.Meta.Block, b)
				ui.Index.Set(index)
				_ = p.Process(PROCESS_DONE, index, start, end, nil)
				_ = s.Save(ui)
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		index++
	}
	if err != nil {
		_ = s.Save(ui)
		_ = p.Exit(err)
		return ui.Meta, err
	}
	ui.Meta.Status = meta.Finish
	ui.Meta.Hash = sh.Sum(nil)
	_ = s.Save(ui)
	_ = s.Finish()
	_ = p.Exit(nil)
	return ui.Meta, nil
}

package storage

import (
	"dxkite.cn/storage/image"
	"dxkite.cn/storage/meta"
	"dxkite.cn/storage/upload"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

type Uploader struct {
	Size int64
	Usn  string
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
	hash := SteamHash(file)
	size := SteamSize(file)
	name := file.Name()
	return u.UploadSteam(name, hash, size, file)
}

func (u *Uploader) UploadSteam(name string, hash []byte, size int64, r io.Reader) error {
	base := filepath.Base(name)
	log.Printf("upload to %s\n", u.Usn)
	log.Printf("upload meta info %x %s %d\n", hash, base, size)

	bc := int64(math.Ceil(float64(size) / float64(u.Size)))
	p := NewFileUploadProcessor(name+EXT_UPLOADING, NewUploadInfo(base, size, bc, u.Size, hash))
	ui, _ := p.Load()

	var index = int64(0)
	var err error
	var buf = make([]byte, u.Size)
	var uploader upload.Uploader

	if u, er := upload.Create(u.Usn); er != nil {
		return er
	} else {
		uploader = u
	}

	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			start, end := IndexToRange(ui.Meta.Size, ui.Meta.BlockSize, index)
			log.Printf("uploading %d/%d block\n", index+1, bc)
			if ui.Index.Get(index) {
				log.Printf("skip uploaded %d block\n", index+1)
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
				Name: fmt.Sprintf("%s-%d.png", hex.EncodeToString(hash), index),
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
				log.Printf("uploaded %d/%d block\n", index+1, bc)
				_ = p.Process(PROCESS_DONE, index, start, end, nil)
				_ = p.Save(ui)
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
		_ = p.Save(ui)
		_ = p.Finish()
		return err
	}
	ui.Meta.Status = meta.Finish
	if er := meta.EncodeToFile(name+EXT_META, ui.Meta); er != nil {
		_ = p.Save(ui)
		_ = p.Finish()
		return er
	}
	log.Println("finished")
	_ = p.Finish()
	return nil
}

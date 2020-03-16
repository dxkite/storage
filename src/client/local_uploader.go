package client

import (
	"dxkite.cn/go-storage/src/bitset"
	"dxkite.cn/go-storage/src/image"
	"dxkite.cn/go-storage/src/meta"
	"dxkite.cn/go-storage/src/upload"
	"dxkite.cn/go-storage/storage"
	"encoding/gob"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type LocalUploader struct {
	Size       int64
	Type       string
	bn         string
	UploadInfo *UploadInfo
}

type UploadInfo struct {
	Index bitset.BitSet
	Meta  *meta.MetaInfo
}

func NewLocalUploader(bs int64, t string) *LocalUploader {
	return &LocalUploader{
		Size: bs,
		Type: t,
	}
}

func (u *LocalUploader) UploadFile(name string) error {
	file, oer := os.OpenFile(name, os.O_RDONLY, os.ModePerm)
	if oer != nil {
		return oer
	}
	var info = SteamHash(file)
	var size = SteamSize(file)
	base := filepath.Base(file.Name())
	log.Printf("upload meta info %x %s %d\n", info, base, size)
	u.bn = hex.EncodeToString(info) + ".uploading"
	var buf = make([]byte, u.Size)
	ui := u.GetUploadInfo(base, size, info)
	u.UploadInfo = ui
	var index = int64(0)
	var err error
	for {
		nr, er := file.Read(buf)
		if nr > 0 {
			log.Printf("uploading %d block\n", index)
			if ui.Index.Get(index) {
				log.Printf("uploaded %d block before\n", index)
				continue
			}
			hh := ByteHash(buf)
			if b, er := image.EncodeByte(buf); er != nil {
				err = er
				break
			} else {
				buf = b
			}
			if r, er := upload.Upload(u.Type, &upload.FileObject{
				Name: strconv.Itoa(int(index)) + ".png",
				Data: buf,
			}); er != nil {
				err = er
				break
			} else {
				ui.Meta.Status = meta.Uploading
				ui.Meta.Block = append(ui.Meta.Block, meta.DataBlock{
					Hash:  hh,
					Index: index,
					Data:  []byte(r.Url),
				})
				ui.Index.Set(index)
				log.Printf("uploaded %d block\n", index)
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
		_ = EncodeUploadInfo(u.bn, ui)
		return err
	}
	ui.Meta.Status = meta.Finish
	if er := meta.EncodeToFile(name+".meta", ui.Meta); er != nil {
		_ = EncodeUploadInfo(u.bn, ui)
		return er
	}
	log.Println("finished")
	return nil
}

func (u *LocalUploader) GetUploadInfo(name string, size int64, info []byte) *UploadInfo {
	if FileExist(u.bn) {
		if ui, er := DecodeUploadInfoFile(u.bn); er != nil {
			return ui
		}
	}
	m := &meta.MetaInfo{
		Hash:      info,
		BlockSize: u.Size,
		Size:      size,
		Name:      name,
		Status:    meta.Create,
		Type:      int32(storage.DataResponse_URI),
		Encode:    int32(storage.DataResponse_IMAGE),
		Block:     []meta.DataBlock{},
	}
	return &UploadInfo{
		Index: bitset.New(size),
		Meta:  m,
	}
}

func EncodeUploadInfo(path string, info *UploadInfo) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewEncoder(f)
	return b.Encode(info)
}

func DecodeUploadInfoFile(path string) (*UploadInfo, error) {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return nil, er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewDecoder(f)
	info := new(UploadInfo)
	der := b.Decode(&info)
	if der != nil {
		return nil, der
	}
	return info, nil
}

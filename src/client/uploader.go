package client

import (
	"dxkite.cn/go-storage/src/bitset"
	"dxkite.cn/go-storage/src/image"
	"dxkite.cn/go-storage/src/meta"
	"dxkite.cn/go-storage/src/upload"
	"encoding/gob"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

type Uploader struct {
	Size       int64
	Type       string
	bn         string
	UploadInfo *UploadInfo
}

type UploadInfo struct {
	Index bitset.BitSet
	Meta  *meta.MetaInfo
}

func NewUploader(bs int64, t string) *Uploader {
	return &Uploader{
		Size: bs,
		Type: t,
	}
}

func (u *Uploader) UploadFile(name string) error {
	file, oer := os.OpenFile(name, os.O_RDONLY, os.ModePerm)
	if oer != nil {
		return oer
	}
	var info = SteamHash(file)
	var size = SteamSize(file)
	base := filepath.Base(file.Name())
	log.Printf("upload meta info %x %s %d\n", info, base, size)
	u.bn = name + ".uploading"
	var buf = make([]byte, u.Size)
	bc := int64(math.Ceil(float64(size) / float64(u.Size)))
	ui := u.GetUploadInfo(base, size, bc, info)
	u.UploadInfo = ui
	var index = int64(0)
	var err error
	for {
		nr, er := file.Read(buf)
		if nr > 0 {
			log.Printf("uploading %d/%d block\n", index+1, bc)
			if ui.Index.Get(index) {
				log.Printf("skip uploaded %d block\n", index+1)
				index++
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
				log.Printf("uploaded %d/%d block\n", index+1, bc)
				_ = EncodeUploadInfo(u.bn, ui)
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
	_ = os.Remove(u.bn)
	return nil
}

func (u *Uploader) GetUploadInfo(name string, size, block int64, info []byte) *UploadInfo {
	if FileExist(u.bn) {
		if ui, er := DecodeUploadInfoFile(u.bn); er == nil {
			return ui
		}
	}
	m := &meta.MetaInfo{
		Hash:      info,
		BlockSize: u.Size,
		Size:      size,
		Name:      name,
		Status:    meta.Create,
		Type:      int32(meta.Type_URI),
		Encode:    int32(meta.Encode_Image),
		Block:     []meta.DataBlock{},
	}
	return &UploadInfo{
		Index: bitset.New(block),
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

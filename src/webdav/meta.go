package webdav

import (
	"dxkite.cn/go-storage/src/meta"
)

func NewLocalMeta(name string, hash []byte, size int64) *meta.Info {
	r := &meta.Info{
		Hash:      hash,
		BlockSize: 0,
		Size:      size,
		Name:      name,
		Status:    meta.Local,
		Type:      int32(meta.Type_URI),
		Encode:    int32(meta.Encode_Image),
		Block:     nil,
	}
	return r
}

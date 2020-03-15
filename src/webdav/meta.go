package webdav

import (
	"dxkite.cn/go-storage/src/meta"
	"dxkite.cn/go-storage/storage"
)

func NewLocalMeta(name string, hash []byte, size int64) *meta.MetaInfo {
	r := &meta.MetaInfo{
		Hash:      hash,
		BlockSize: 0,
		Size:      size,
		Name:      name,
		Status:    meta.Local,
		Type:      int32(storage.DataResponse_STREAM),
		Encode:    int32(storage.DataResponse_NONE),
		Block:     nil,
	}
	return r
}

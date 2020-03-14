package client

import (
	"dxkite.cn/go-storage/src/meta"
	"dxkite.cn/go-storage/storage"
)

func NewMeta(m *storage.DataResponse) *meta.MetaInfo {
	r := &meta.MetaInfo{
		Hash:      m.Hash,
		BlockSize: m.Block,
		Size:      m.Size,
		Name:      m.Name,
		Type:      int32(storage.DataResponse_URI),
		Encode:    int32(storage.DataResponse_IMAGE),
		Block:     CreateBlocks(m.Blocks),
	}
	return r
}

func CreateBlocks(b []*storage.DataBlock) []meta.DataBlock {
	bb := []meta.DataBlock{}
	for _, ib := range b {
		bb = append(bb, meta.DataBlock{
			Hash:  ib.Hash,
			Index: ib.Index,
			Data:  ib.Data,
		})
	}
	return bb
}

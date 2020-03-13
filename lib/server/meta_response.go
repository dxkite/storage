package server

import (
	"dxkite.cn/go-storage/lib/meta"
	"dxkite.cn/go-storage/storage"
)

func NewDataResponse(m *meta.MetaInfo) *storage.DataResponse {
	if m == nil {
		return &storage.DataResponse{}
	}
	r := &storage.DataResponse{
		Hash:   m.Hash,
		Block:  int64(len(m.Block)),
		Size:   m.Size,
		Type:   storage.DataResponse_URI,
		Encode: storage.DataResponse_IMAGE,
		Blocks: CreateBlocks(m.Block),
	}
	return r
}

func CreateBlocks(b []meta.DataBlock) []*storage.DataBlock {
	bb := []*storage.DataBlock{}
	for _, ib := range b {
		bb = append(bb, &storage.DataBlock{
			Hash:  ib.Hash,
			Index: ib.Index,
			Data:  ib.Data,
		})
	}
	return bb
}

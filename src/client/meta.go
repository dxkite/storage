package client

import (
	"dxkite.cn/go-storage/src/bitset"
	"dxkite.cn/go-storage/src/meta"
	"encoding/gob"
	"os"
)

type DownloadMeta struct {
	meta.MetaInfo
	Index         bitset.BitSet
	Downloaded    int
	DownloadTotal int
}

func EncodeToFile(path string, info *DownloadMeta) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewEncoder(f)
	return b.Encode(info)
}

func DecodeToFile(path string) (*DownloadMeta, error) {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return nil, er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewDecoder(f)
	info := new(DownloadMeta)
	der := b.Decode(&info)
	if der != nil {
		return nil, der
	}
	return info, nil
}

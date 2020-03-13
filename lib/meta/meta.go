package meta

import (
	"encoding/gob"
	"os"
)

type MetaInfo struct {
	Hash   []byte
	Size   int64
	Encode int32
	Blocks []DataBlock
}

type DataBlock struct {
	Hash  []byte
	Type  int32
	Index int64
	Data  []byte
}

func EncodeToFile(path string, info *MetaInfo) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if er != nil {
		return er
	}
	b := gob.NewEncoder(f)
	return b.Encode(info)
}

func DecodeToFile(path string) (*MetaInfo, error) {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return nil, er
	}
	b := gob.NewDecoder(f)
	info := new(MetaInfo)
	der := b.Decode(&info)
	if der != nil {
		return nil, der
	}
	return info, nil
}

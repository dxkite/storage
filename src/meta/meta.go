package meta

import (
	"encoding/gob"
	"os"
)

const (
	Create int = iota
	Uploading
	Finish
)

type MetaInfo struct {
	Status    int
	Hash      []byte
	Name      string
	Size      int64
	BlockSize int64
	Encode    int32
	Type      int32
	Block     []DataBlock
}

type DataBlock struct {
	Hash  []byte
	Index int64
	Data  []byte
}

func (m *MetaInfo) AppendBlock(b *DataBlock) {
	if m.Block == nil {
		m.Block = []DataBlock{*b}
		return
	}
	for i, bb := range m.Block {
		if bb.Index == b.Index {
			m.Block[i] = *b
			return
		}
	}
	m.Block = append(m.Block, *b)
}

func EncodeToFile(path string, info *MetaInfo) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
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
	defer func() { _ = f.Close() }()
	return info, nil
}

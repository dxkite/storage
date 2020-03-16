package meta

import (
	"github.com/zeebo/bencode"
	"os"
)

type Status int
type EncodeType int
type DataType int

const (
	Create Status = iota
	Uploading
	Finish
	Local
)

const (
	Encode_None EncodeType = iota
	Encode_Image
)
const (
	Type_URI DataType = iota
	Type_Stream
)

type Info struct {
	Status    Status      `bencode:"status"`
	Hash      []byte      `bencode:"hash"`
	Name      string      `bencode:"name"`
	Size      int64       `bencode:"size"`
	BlockSize int64       `bencode:"block_size"`
	Encode    int32       `bencode:"encode"`
	Type      int32       `bencode:"type"`
	Block     []DataBlock `bencode:"blocks"`
}

type DataBlock struct {
	Hash  []byte `bencode:"hash"`
	Index int64  `bencode:"index"`
	Data  []byte `bencode:"data"`
}

func (m *Info) AppendBlock(b *DataBlock) {
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

func EncodeToFile(path string, info *Info) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := bencode.NewEncoder(f)
	return b.Encode(info)
}

func DecodeFromFile(path string) (*Info, error) {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return nil, er
	}
	b := bencode.NewDecoder(f)
	info := new(Info)
	der := b.Decode(&info)
	if der != nil {
		return nil, der
	}
	defer func() { _ = f.Close() }()
	return info, nil
}

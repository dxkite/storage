package meta

import (
	"errors"
	"github.com/zeebo/bencode"
	"io"
	"math/rand"
	"os"
	"time"
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

const Magic = "\x14SMF"
const Version = 1

type Info struct {
	Status    Status      `bencode:"status"`
	Hash      []byte      `bencode:"hash"`
	Name      string      `bencode:"name"`
	Size      int64       `bencode:"size"`
	BlockSize int64       `bencode:"block_size"`
	Encode    int32       `bencode:"encode"`
	Type      int32       `bencode:"type"`
	Block     []DataBlock `bencode:"block"`
}

type DataBlock struct {
	Hash  []byte `bencode:"h"`
	Index int64  `bencode:"i"`
	Data  []byte `bencode:"d"`
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
	return EncodeToStream(f, info)
}

func EncodeToStream(f io.Writer, info *Info) error {
	rand.Seed(time.Now().Unix())
	xor := byte(rand.Intn(254) + 1)
	if _, er := f.Write([]byte(Magic)); er != nil {
		return er
	}
	if _, er := f.Write([]byte{Version, xor}); er != nil {
		return er
	}
	b := bencode.NewEncoder(NewXORWriter(xor, f))
	return b.Encode(info)
}

func DecodeFromStream(f io.Reader) (*Info, error) {
	var bb = make([]byte, 6)
	var xor byte
	if _, er := f.Read(bb); er != nil {
		return nil, er
	} else {
		if magic := bb[0:4]; string(magic) != Magic {
			return nil, errors.New("unknown magic header")
		}
		xor = bb[5]
	}
	b := bencode.NewDecoder(NewXORReader(xor, f))
	info := new(Info)
	der := b.Decode(&info)
	if der != nil {
		return nil, errors.New("bencode: " + der.Error())
	}
	return info, nil
}

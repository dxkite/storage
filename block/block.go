package block

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	errorSeek = errors.New("error seek")
)

type File interface {
	io.ReadWriteSeeker
	io.Closer
}

type BlockFile struct {
	File File   // 文件
	Hash []byte // 文件HASH
	lock sync.Mutex
}

type block struct {
	start int64
	end   int64
	bytes []byte
}

type Block interface {
	Start() int64
	End() int64
	Bytes() []byte
}

func NewBlock(start, end int64, data []byte) Block {
	return &block{
		start: start,
		end:   end,
		bytes: data,
	}
}

func (b *block) Start() int64 {
	return b.start
}
func (b *block) End() int64 {
	return b.end
}
func (b *block) Bytes() []byte {
	return b.bytes
}

func (b *BlockFile) WriteBlock(c Block) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	l := c.End() - c.Start()
	if l > int64(len(c.Bytes())) {
		return io.ErrShortBuffer
	}
	if s, err := b.File.Seek(c.Start(), io.SeekStart); err != nil {
		return err
	} else {
		if s != c.Start() {
			return errorSeek
		}
	}
	if n, err := b.File.Write(c.Bytes()[:l]); err != nil {
		return err
	} else {
		if int64(n) != l {
			return io.ErrShortWrite
		}
	}
	return nil
}

func (b *BlockFile) ReadBlock(c Block) ([]byte, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	l := c.End() - c.Start()
	buf := make([]byte, l)
	if s, err := b.File.Seek(c.Start(), io.SeekStart); err != nil {
		return nil, err
	} else {
		if s != c.Start() {
			return nil, errorSeek
		}
	}
	if nr, er := b.File.Read(buf); er != nil {
		return nil, er
	} else {
		if int64(nr) != l {
			return nil, errors.New("error read size")
		}
		return buf, nil
	}
}

func (b *BlockFile) CheckSum() bool {
	h := sha1.New()
	_, _ = b.File.Seek(0, io.SeekStart)
	_, err := io.Copy(h, b.File)
	if err != nil {
		panic(fmt.Sprintf("check sum: %v", err))
	}
	return bytes.Compare(h.Sum(nil), b.Hash) == 0
}

func (b *BlockFile) Close() error {
	return b.File.Close()
}

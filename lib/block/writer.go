package block

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
)

var (
	errorSeek = errors.New("error seek")
)

type BlockFile struct {
	file io.ReadWriteSeeker // 文件
	Hash []byte             // 文件HASH
}

type block struct {
	start int64
	end   int64
	data  []byte
}

type Block interface {
	Start() int64
	End() int64
	Data() []byte
}

func NewBlock(start, end int64, data []byte) Block {
	return &block{
		start: start,
		end:   end,
		data:  data,
	}
}

func (b *block) Start() int64 {
	return b.start
}
func (b *block) End() int64 {
	return b.end
}
func (b *block) Data() []byte {
	return b.data
}

func (b *BlockFile) WriteBlock(c Block) error {
	l := c.End() - c.Start()
	if l != int64(len(c.Data())) {
		return io.ErrShortBuffer
	}
	if s, err := b.file.Seek(int64(c.Start()), io.SeekStart); err != nil {
		return err
	} else {
		if int64(s) != c.Start() {
			return errorSeek
		}
	}
	if n, err := b.file.Write(c.Data()[:l]); err != nil {
		return err
	} else {
		if int64(n) != l {
			return io.ErrShortWrite
		}
	}
	return nil
}

func (b *BlockFile) CheckSum() bool {
	h := sha1.New()
	_, _ = b.file.Seek(0, io.SeekStart)
	_, err := io.Copy(h, b.file)
	if err != nil {
		panic(fmt.Sprintf("check sum: %v", err))
	}
	return bytes.Compare(h.Sum(nil), b.Hash) == 0
}

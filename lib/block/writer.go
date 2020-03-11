package block

import (
	"bytes"
	"crypto/sha1"
	"dxkite.cn/go-storage/storage"
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

func (b *BlockFile) WriteRange(r storage.ContentRange, buf []byte) error {
	l := r.End - r.Start
	if l != uint64(len(buf)) {
		return io.ErrShortBuffer
	}
	if s, err := b.file.Seek(int64(r.Start), io.SeekStart); err != nil {
		return err
	} else {
		if uint64(s) != r.Start {
			return errorSeek
		}
	}

	if n, err := b.file.Write(buf[:l]); err != nil {
		return err
	} else {
		if uint64(n) != l {
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

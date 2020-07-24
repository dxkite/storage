package storage

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) { // 根据错误类型进行判断
			return true
		}
		return false
	}
	return true
}

// 计算块位置
func IndexToRange(size, blockSize, index int64) (begin, end int64) {
	begin = index * blockSize
	end = begin + blockSize
	if end > size {
		end = size
	}
	return begin, end
}

func SteamHash(r io.ReadSeeker) []byte {
	h := sha1.New()
	_, _ = r.Seek(0, io.SeekStart)
	_, err := io.Copy(h, r)
	if err != nil {
		panic(fmt.Sprintf("check sum: %v", err))
	}
	_, _ = r.Seek(0, io.SeekStart)
	return h.Sum(nil)
}

func ByteHash(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	return h.Sum(nil)
}

func SteamSize(r io.ReadSeeker) int64 {
	n, _ := r.Seek(0, io.SeekEnd)
	_, _ = r.Seek(0, io.SeekStart)
	return n
}

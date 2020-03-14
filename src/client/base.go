package client

import (
	"crypto/sha1"
	"fmt"
	"io"
)

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

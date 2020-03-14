package webdav

import (
	"dxkite.cn/go-storage/src/meta"
	"io"
	"os"
)

type ReadWriteCloseSeeker interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

type RemoteFile struct {
	MetaPath string
	Local    string
	Meta     *meta.MetaInfo
}

func prepareMetaStream(file File) {
	r := RealName(file.Path)
	if FileExist(r) {
		f, err := os.OpenFile(r, os.O_RDWR, os.ModePerm)
		if err == nil {
			file.real = f
		}
	}
}

func (f RemoteFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f RemoteFile) Read(p []byte) (n int, err error) {
	return n, nil
}

func (f RemoteFile) Close() error {
	return nil
}

func (f RemoteFile) Write(p []byte) (n int, err error) {
	return 0, nil
}

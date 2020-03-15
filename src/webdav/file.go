package webdav

import (
	"dxkite.cn/go-storage/src/client"
	"dxkite.cn/go-storage/src/meta"
	"encoding/hex"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ReadWriteCloseSeeker interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

type RemoteFile struct {
	webdav.File
	Path  string
	Index *IndexFile
	Meta  *meta.MetaInfo

	fs      FileSystem
	newFile bool
	tmp     string
	object  ReadWriteCloseSeeker
}

func NewRemoteFile(d FileSystem, name, tmp string, file webdav.File) *RemoteFile {
	newFile := len(tmp) > 0
	if newFile == false && IsDir(name) == false {
		log.Println("NewRemoteFile", name, "DecodeIndexFile")
		// 索引文件存在
		i, er := DecodeIndexFile(name)
		if er == nil {
			m, _ := meta.DecodeFromFile(i.GetMeta(d))
			object, _ := os.OpenFile(i.GetObject(d), os.O_CREATE|os.O_RDWR, os.ModePerm)

			return &RemoteFile{
				fs:      d,
				tmp:     tmp,
				newFile: newFile,

				File:   file,
				Index:  i,
				Path:   name,
				object: object,
				Meta:   m,
			}
		}
	}

	log.Println("NewRemoteFile", name, tmp)
	return &RemoteFile{
		fs:      d,
		tmp:     tmp,
		newFile: newFile,

		File: file,
		Path: name,
	}
}

func (f *RemoteFile) Readdir(count int) ([]os.FileInfo, error) {
	readdir, err := f.File.Readdir(count)
	if err != nil {
		return nil, err
	}
	list := []os.FileInfo{}
	for _, item := range readdir {
		p := path.Join(f.Path, item.Name())
		if item.Mode().IsRegular() {
			if strings.HasSuffix(item.Name(), IndexSuffix) {
				list = append(list, WrapIndexFile(f.fs, p, item))
			}
		} else {
			list = append(list, item)
		}
	}
	return list, err
}

func (f *RemoteFile) Write(p []byte) (n int, err error) {
	if f.newFile {
		log.Println("Write is New File")
	}
	if f.object != nil {
		return f.object.Write(p)
	}
	return f.File.Write(p)
}

func (f *RemoteFile) Close() error {
	if f.object != nil {
		_ = f.File.Close()
		return f.object.Close()
	}
	if f.newFile {
		hash := client.SteamHash(f.File)
		size := client.SteamSize(f.File)
		_ = f.File.Close()
		log.Println("upload file", f.Path, "success")
		return f.createNewFile(hash, size)
	}
	return f.File.Close()
}

func (f *RemoteFile) createNewFile(hash []byte, size int64) error {
	name := indexNameToRealName(filepath.Base(f.Path))
	m := NewLocalMeta(name, hash, size)
	hh := hex.EncodeToString(hash)
	i := NewIndexFile(hh)
	ofn := i.GetObject(f.fs)
	mfn := i.GetMeta(f.fs)
	ff := f.tmp

	// 复制文件
	if er := Copy(ff, ofn); er != nil {
		_ = os.Remove(ff)
		log.Println("copy object error", er)
		return er
	}
	// 删除文件
	if er := os.Remove(ff); er != nil {
		_ = os.Remove(ff)
		log.Println("remove object error", er)
		return er
	}
	log.Println("Move", ff, ofn)

	// 重写元数据
	if er := meta.EncodeToFile(mfn, m); er != nil {
		// 移动失败则删除
		_ = os.Remove(ff)
		log.Println("encode meta error", er)
		return er
	}

	// 重写index
	ii := NewIndexFile(hex.EncodeToString(hash))
	if er := EncodeIndexFile(f.Path, &ii); er != nil {
		// 移动失败则删除
		_ = os.Remove(ff)
		log.Println("encode index error", er)
		return er
	}
	log.Println("NewIndexFile", f.Path)
	return nil
}

func (f *RemoteFile) Read(p []byte) (n int, err error) {
	if f.object != nil {
		return f.object.Write(p)
	}
	return f.File.Read(p)
}

func (f *RemoteFile) Seek(offset int64, whence int) (int64, error) {
	if f.object != nil {
		return f.object.Seek(offset, whence)
	}
	return f.File.Seek(offset, whence)
}

func (f *RemoteFile) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	return WrapIndexFile(f.fs, f.Path, fi), err
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func Copy(src, dst string) (err error) {
	var fs, fd *os.File
	if fs, err = os.OpenFile(src, os.O_RDONLY, os.ModePerm); err != nil {
		return
	}
	if fd, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm); err != nil {
		return
	}
	defer func() { _ = fd.Close() }()
	defer func() { _ = fs.Close() }()
	if _, er := io.Copy(fd, fs); er != nil {
		if er != io.EOF {
			return er
		}
	}
	return nil
}

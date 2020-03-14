package webdav

import (
	"dxkite.cn/go-storage/src/meta"
	"golang.org/x/net/webdav"
	"os"
	"path"
	"strings"
)

const MetaSuffix = ".meta"

type File struct {
	webdav.File
	Path   string
	IsMeta bool
	Exist  bool
	real   ReadWriteCloseSeeker
}

type MetaFileInfo struct {
	os.FileInfo
	name string
	Meta *meta.MetaInfo
}

func (f File) Readdir(count int) ([]os.FileInfo, error) {
	readdir, err := f.File.Readdir(count)
	if err != nil {
		return nil, err
	}
	//log.Println("Readdir", f.Path)
	list := []os.FileInfo{}
	for _, item := range readdir {
		if IsMetaFile(item) {
			p := path.Join(f.Path, item.Name())
			//log.Println(f.Path, "Real Name", RealName(p), "FileExist", FileExist(RealName(p)))
			if FileExist(RealName(p)) == false {
				m := WrapMetaFile(p, item)
				list = append(list, m)
			}
		} else {
			list = append(list, item)
		}
	}
	return list, err
}

func IsMetaFile(f os.FileInfo) bool {
	return f.Mode().IsRegular() && strings.HasSuffix(f.Name(), MetaSuffix)
}

func RealName(name string) string {
	return name[0 : len(name)-len(MetaSuffix)]
}

func WrapMetaFile(path string, f os.FileInfo) os.FileInfo {
	m, _ := meta.DecodeFromFile(path)
	if m != nil {
		n := RealName(f.Name())
		//log.Println("wrap meta", f.Name(), "to", n)
		return MetaFileInfo{
			FileInfo: f,
			name:     n,
			Meta:     m,
		}
	}
	return f
}

func NewFile(file webdav.File, path string, isMeta bool, exist bool) File {
	f := File{
		File:   file,
		Path:   path,
		IsMeta: isMeta,
		Exist:  exist,
		real:   nil,
	}
	if isMeta {
		prepareMetaStream(&f, exist)
	}
	return f
}

func (f File) Write(p []byte) (n int, err error) {
	//log.Println("Write f.real", f.real, f.Path)
	if f.real != nil {
		return f.real.Write(p)
	}
	return f.File.Write(p)
}

func (f File) Close() error {
	//log.Println("Close f.real", f.real, f.Path)
	if f.real != nil {
		return f.real.Close()
	}
	return f.File.Close()
}

func (f File) Read(p []byte) (n int, err error) {
	//log.Println("Read f.real", f.real, f.Path)
	if f.real != nil {
		return f.real.Read(p)
	}
	return f.File.Read(p)
}

func (f File) Seek(offset int64, whence int) (int64, error) {
	//log.Println("Seek f.real", f.real, f.Path)
	if f.real != nil {
		//log.Println("Seek real", f.Path, offset, whence)
		return f.real.Seek(offset, whence)
	}
	return f.File.Seek(offset, whence)
}

func (f File) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	if f.IsMeta {
		return WrapMetaFile(f.Path, fi), nil
	}
	return fi, err
}

func (f MetaFileInfo) Name() string {
	return f.name
}

func (f MetaFileInfo) Size() int64 {
	return f.Meta.Size
}

func (MetaFileInfo) Sys() interface{} {
	return nil
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

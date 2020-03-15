package webdav

import (
	"dxkite.cn/go-storage/src/meta"
	"os"
)

type RemoteFileInfo struct {
	os.FileInfo
	name string
	Meta *meta.MetaInfo
}

func (f RemoteFileInfo) Name() string {
	return f.name
}

func (f RemoteFileInfo) Size() int64 {
	return f.Meta.Size
}

func (RemoteFileInfo) Sys() interface{} {
	return nil
}

func WrapIndexFile(fs FileSystem, p string, f os.FileInfo) os.FileInfo {
	i, er := DecodeIndexFile(p)
	if er == nil {
		n := indexNameToRealName(f.Name())
		if m, err := meta.DecodeFromFile(i.GetMeta(fs)); err == nil {
			return RemoteFileInfo{
				FileInfo: f,
				name:     n,
				Meta:     m,
			}
		}
	}
	return f
}

func indexNameToRealName(name string) string {
	return name[0 : len(name)-len(IndexSuffix)]
}

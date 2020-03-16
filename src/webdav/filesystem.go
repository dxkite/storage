package webdav

import (
	"context"
	"dxkite.cn/go-storage/src/client"
	"encoding/hex"
	"golang.org/x/net/webdav"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FileSystem struct {
	webdav.Dir
	MetaRoot   string
	ObjectRoot string
	UploadRoot string
}

func NewSystem(root string) FileSystem {
	idx := path.Join(root, ".storage", "index")
	fs := FileSystem{
		Dir:        webdav.Dir(path.Join(root, ".storage", "index")),
		MetaRoot:   path.Join(root, ".storage", "meta"),
		ObjectRoot: path.Join(root, ".storage", "object"),
		UploadRoot: path.Join(root, ".storage", "upload"),
	}
	_ = os.MkdirAll(idx, os.ModePerm)
	_ = os.MkdirAll(fs.MetaRoot, os.ModePerm)
	_ = os.MkdirAll(fs.ObjectRoot, os.ModePerm)
	_ = os.MkdirAll(fs.UploadRoot, os.ModePerm)
	return fs
}

func (d FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	n, exist := d.getNameIndex(name)
	var t string
	var rf string

	if exist {
		// 打开存在文件
		name = d.resolve(n)
		log.Println("open file", name)
		rf = name
	} else {
		// 打开一个不存在的文件
		nn := hex.EncodeToString(client.ByteHash([]byte(name)))
		t = path.Join(d.UploadRoot, nn)
		name = d.resolve(n)
		log.Println("new file", n, "real", name, "save", t)
		rf = t
	}

	f, err := os.OpenFile(rf, flag, perm)
	if err != nil {
		return nil, err
	}
	return NewRemoteFile(d, name, t, f), nil
}

func (d FileSystem) RemoveAll(ctx context.Context, name string) error {
	metaName, _ := d.getNameIndex(name)
	return d.Dir.RemoveAll(ctx, metaName)
}

func (d FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	oldMetaName, _ := d.getNameIndex(oldName)
	newName = newName + IndexSuffix
	return d.Dir.Rename(ctx, oldMetaName, newName)
}

func (d FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name, _ = d.getNameIndex(name)
	return d.Dir.Stat(ctx, name)
}

func (d FileSystem) getNameIndex(name string) (string, bool) {
	p := d.resolve(name)
	if IsDir(p) {
		return name, true
	}
	return name + IndexSuffix, FileExist(p + IndexSuffix)
}

func (d FileSystem) resolve(name string) string {
	// This implementation is based on Dir.Open's code in the standard net/http package.
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return ""
	}
	dir := string(d.Dir)
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, filepath.FromSlash(slashClean(name)))
}

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

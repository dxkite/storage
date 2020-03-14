package webdav

import (
	"context"
	"golang.org/x/net/webdav"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem struct {
	webdav.Dir
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

func (d FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name, _ = d.getRealName(name)
	p := d.resolve(name)
	f, err := d.Dir.OpenFile(ctx, name, flag, perm)
	if err != nil {
		return nil, err
	}
	return File{File: f, Path: p}, nil
}

func (d FileSystem) RemoveAll(ctx context.Context, name string) error {
	name, _ = d.getRealName(name)
	return d.Dir.RemoveAll(ctx, name)
}

func (d FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	var isMeta = false
	oldName, isMeta = d.getRealName(oldName)
	if isMeta {
		newName = newName + MetaSuffix
	}
	return d.Dir.Rename(ctx, oldName, newName)
}

func (d FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name, _ = d.getRealName(name)
	//p := d.resolve(name)
	return d.Dir.Stat(ctx, name)
}

func (d FileSystem) getRealName(name string) (string, bool) {
	p := d.resolve(name)
	if FileExist(p) {
		return name, false
	}
	if FileExist(p + ".meta") {
		return name + ".meta", true
	}
	return p, false
}

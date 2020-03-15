package webdav

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

const MetaSuffix = ".meta"
const UploadSuffix = "-"
const IndexSuffix = ".index"
const ObjectSuffix = ".object"

type IndexFile struct {
	Hash string
}

func NewIndexFile(hash string) IndexFile {
	return IndexFile{Hash: hash}
}

func (i IndexFile) GetMeta(fs FileSystem) string {
	return path.Join(fs.MetaRoot, i.Hash+MetaSuffix)
}

func (i IndexFile) GetObject(fs FileSystem) string {
	return path.Join(fs.ObjectRoot, i.Hash+ObjectSuffix)
}

func EncodeIndexFile(path string, info *IndexFile) error {
	return ioutil.WriteFile(path, []byte(info.Hash), os.ModePerm)
}

func DecodeIndexFile(path string) (*IndexFile, error) {
	if b, er := ioutil.ReadFile(path); er == nil {
		if len(string(b)) != 40 {
			return nil, errors.New("error index hash")
		}
		h := NewIndexFile(string(b))
		return &h, nil
	} else {
		return nil, er
	}
}

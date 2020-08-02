package upload

import (
	"errors"
	"net/url"
	"strings"
)

type FileObject struct {
	// 文件名
	Name string
	// 文件内容
	Data []byte
}

type Config map[string]string

type Result struct {
	// 上传URL
	Url string
	// 原始数据
	Raw []byte
}

var Default = OppoFeedback

var list = make(map[string]UploaderCreator)

type Uploader interface {
	Upload(object *FileObject) (*Result, error)
}

type UploaderCreator func(config Config) Uploader

func Register(name string, uploader UploaderCreator) {
	list[name] = uploader
}

func Upload(usn string, object *FileObject) (*Result, error) {
	if uploader, err := Create(usn); err == nil {
		return uploader.Upload(object)
	} else {
		return nil, err
	}
}

func Create(usn string) (Uploader, error) {
	if name, cfg, err := parse(usn); err == nil {
		return CreateConfig(name, cfg)
	} else {
		return nil, err
	}
}

func CreateConfig(name string, config Config) (Uploader, error) {
	if creator, ok := list[name]; ok {
		return creator(config), nil
	}
	return nil, errors.New("unknown server type:" + name)
}

func With(usn string) Uploader {
	if uploader, err := Create(usn); err != nil {
		panic(err)
	} else {
		return uploader
	}
}

func WithConfig(name string, config Config) Uploader {
	if uploader, err := CreateConfig(name, config); err != nil {
		panic(err)
	} else {
		return uploader
	}
}

func parse(usn string) (name string, cfg Config, err error) {
	n, q := split(usn, ':')
	name = n
	// name:uid=xxx
	if u, per := url.ParseQuery(q); per != nil {
		return "", nil, errors.New("parser upload server name error:" + per.Error())
	} else {
		cfg = map[string]string{}
		for name, value := range u {
			cfg[name] = value[0]
		}
		return name, cfg, nil
	}
}

func split(s string, sep byte) (string, string) {
	i := strings.IndexByte(s, sep)
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

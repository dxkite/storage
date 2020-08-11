package upload

import (
	"bytes"
	"errors"
	"net/url"
	"os/exec"
	"strings"
)

const PLUGIN = "plugin"

type PluginUploader struct {
	config Config
}

func init() {
	Register(PLUGIN, func(config Config) Uploader {
		return &PluginUploader{
			config: config,
		}
	})
}

func (p *PluginUploader) Upload(object *FileObject) (*Result, error) {
	if p.config.Host == "cmd" {
		return p.UploadByCmd(object)
	}
	return nil, errors.New("unknown plugin type")
}

func (p *PluginUploader) UploadByCmd(object *FileObject) (*Result, error) {
	ret := &bytes.Buffer{}
	var val url.Values
	if q, er := url.ParseQuery(p.config.RawQuery); er != nil {
		return nil, er
	} else {
		val = q
	}
	cmd := exec.Command(val.Get("exec"))
	cmd.Stdin = bytes.NewBuffer(object.Data)
	cmd.Stdout = ret
	if err := cmd.Run(); err != nil {
		return nil, err
	} else {
		return &Result{
			Url: strings.TrimSpace(ret.String()),
			Raw: ret.Bytes(),
		}, nil
	}
}

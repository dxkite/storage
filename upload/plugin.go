package upload

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

const PLUGIN = "plugin"

type PluginUploader struct {
	conf Config
}

func init() {
	Register(PLUGIN, func(config Config) Uploader {
		return &PluginUploader{
			conf: config,
		}
	})
}

func (p *PluginUploader) Upload(object *FileObject) (*Result, error) {
	if p.conf.Host == "cmd" {
		return p.UploadByCmd(object)
	}
	return nil, errors.New("unknown plugin type")
}

func (p *PluginUploader) UploadByCmd(object *FileObject) (*Result, error) {
	ret := &bytes.Buffer{}
	ers := &bytes.Buffer{}
	val, _ := url.ParseQuery(p.conf.RawQuery)
	cmd := exec.Command(val.Get("exec"), val["args"]...)
	cmd.Stdin = bytes.NewBuffer(object.Data)
	cmd.Stdout = ret
	cmd.Stderr = ers
	if err := cmd.Run(); err != nil {
		return nil, errors.New(fmt.Sprintf("%v: %s", err, ers))
	} else {
		return &Result{
			Url: strings.TrimSpace(ret.String()),
			Raw: ret.Bytes(),
		}, nil
	}
}

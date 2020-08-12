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
	ob := &bytes.Buffer{}
	eb := &bytes.Buffer{}
	q, _ := url.ParseQuery(p.conf.RawQuery)
	cmd := exec.Command(q.Get("exec"), q["args"]...)
	cmd.Stdin = bytes.NewBuffer(object.Data)
	cmd.Stdout = ob
	cmd.Stderr = eb
	if err := cmd.Run(); err != nil {
		return nil, errors.New(fmt.Sprintf("exec error: %v\nstderr:%s\nstdout:%s", err, eb.String(), ob.String()))
	} else {
		return &Result{
			Url: strings.TrimSpace(ob.String()),
			Raw: ob.Bytes(),
		}, nil
	}
}

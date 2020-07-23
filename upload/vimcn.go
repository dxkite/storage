package upload

import (
	"bytes"
	"dxkite.cn/storage/common"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

const VIM_CN = "vim-cn"

type VimCNUploader struct {
}

func init() {
	Register(VIM_CN, func(config Config) Uploader {
		return &VimCNUploader{}
	})
}

func (*VimCNUploader) Upload(object *FileObject) (*Result, error) {
	url := "https://img.vim-cn.com/"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if fw, e := w.CreateFormFile("file", object.Name); e == nil && fw != nil {
		if _, er := fw.Write(object.Data); er != nil {
			return nil, er
		}
	}

	if er := w.Close(); er != nil {
		return nil, er
	}

	req, _ := http.NewRequest(http.MethodPost, url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, er := common.Client.Do(req)
	if er != nil {
		return nil, errors.New(fmt.Sprintf("request error: %v", er))
	}

	defer res.Body.Close()
	body, rer := ioutil.ReadAll(res.Body)

	if rer != nil {
		return nil, errors.New(fmt.Sprintf("read body error: %v", rer))
	}

	if res.StatusCode != 200 {
		return nil, errors.New("vim-cn status error:" + string(body))
	}

	return &Result{
		Url: strings.TrimSpace(string(body)),
		Raw: body,
	}, nil
}

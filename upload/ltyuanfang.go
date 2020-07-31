package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const YUAN_FANG = "yuan-fang"

type YFUploader struct {
}

type YFData struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type YFResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Time int    `json:"time"`
	Data YFData `json:"data"`
}

func init() {
	Register(YUAN_FANG, func(config Config) Uploader {
		return &YFUploader{}
	})
}

func (*YFUploader) Upload(object *FileObject) (*Result, error) {
	url := "https://tc.ltyuanfang.cn/api/upload"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if fw, e := w.CreateFormFile("image", object.Name); e == nil && fw != nil {
		if _, er := fw.Write(object.Data); er != nil {
			return nil, er
		}
	}

	if er := w.Close(); er != nil {
		return nil, er
	}

	req, _ := http.NewRequest(http.MethodPost, url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, er := Client.Do(req)
	if er != nil {
		return nil, errors.New(fmt.Sprintf("request error: %v", er))
	}

	defer func() { _ = res.Body.Close() }()
	body, rer := ioutil.ReadAll(res.Body)

	if rer != nil {
		return nil, errors.New(fmt.Sprintf("read body error: %v", rer))
	}

	if res.StatusCode != 200 {
		return nil, errors.New("vim-cn status error:" + string(body))
	}

	resp := new(YFResponse)
	if er := json.Unmarshal(body, resp); er == nil {
		if resp.Code != 200 {
			return nil, errors.New("YFUploader upload error: " + resp.Msg)
		}
		return &Result{
			Url: resp.Data.Url,
			Raw: body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("decode body error: %v: %s", er, string(body)))
	}
}

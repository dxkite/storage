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

type NetEasy163Response struct {
	Status string   `json:"code"`
	Data   []string `json:"data"`
}

const NETEASY_163 = "163"
const NETEASY = NETEASY_163

type NetEasy163Uploader struct {
}

func init() {
	Register(NETEASY, func(config Config) Uploader {
		return &NetEasy163Uploader{}
	})
}

func (*NetEasy163Uploader) Upload(object *FileObject) (*Result, error) {
	url := "http://you.163.com/xhr/file/upload.json"
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
	res, er := Client.Do(req)
	if er != nil {
		return nil, errors.New(fmt.Sprintf("request error: %v", er))
	}

	defer res.Body.Close()
	body, rer := ioutil.ReadAll(res.Body)
	if rer != nil {
		return nil, errors.New(fmt.Sprintf("read body error: %v", rer))
	}

	resp := new(NetEasy163Response)
	if er := json.Unmarshal(body, resp); er == nil {
		if resp.Status != "200" {
			return nil, errors.New("NetEasy163 upload error: " + string(body))
		}
		return &Result{
			Url: resp.Data[0],
			Raw: body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("decode body error: %v: %s", er, string(body)))
	}
}

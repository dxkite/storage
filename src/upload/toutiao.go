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

type ToutiaoResponse struct {
	Status string `json:"message"`
	Url    string `json:"web_url"`
}

const TOUTIAO = "toutiao"

type TouTiaoUploader struct {
}

func init() {
	Register(TOUTIAO, func(config Config) Uploader {
		return &TouTiaoUploader{}
	})
}

func (*TouTiaoUploader) Upload(object *FileObject) (*Result, error) {
	url := "https://mp.toutiao.com/upload_photo/?type=json"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if fw, e := w.CreateFormFile("photo", object.Name); e == nil && fw != nil {
		if _, er := fw.Write(object.Data); er != nil {
			return nil, er
		}
	}

	if er := w.Close(); er != nil {
		return nil, er
	}

	req, _ := http.NewRequest(http.MethodPost, url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, er := http.DefaultClient.Do(req)
	if er != nil {
		return nil, errors.New(fmt.Sprintf("request error: %v", er))
	}

	defer res.Body.Close()
	body, rer := ioutil.ReadAll(res.Body)
	if rer != nil {
		return nil, errors.New(fmt.Sprintf("read body error: %v", rer))
	}

	resp := new(ToutiaoResponse)
	if er := json.Unmarshal(body, resp); er == nil {
		if resp.Status != "success" {
			return nil, errors.New("TouTiao upload error: " + string(body))
		}
		return &Result{
			Url: resp.Url,
			Raw: body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("decode body error: %v: %s", er, string(body)))
	}
}

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

type DataItem struct {
	Url    string `json:"key"`
	Domain string `json:"domain"`
}

type JueJinResponse struct {
	Status string   `json:"m"`
	Data   DataItem `json:"d"`
}

const JUEJIN = "juejin"

type JueJinUploader struct {
}

func init() {
	Register(JUEJIN, func(config Config) Uploader {
		return &JueJinUploader{}
	})
}

func (*JueJinUploader) Upload(object *FileObject) (*Result, error) {
	url := "https://cdn-ms.juejin.im/v1/upload?bucket=gold-user-assets"
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
	res, er := http.DefaultClient.Do(req)
	if er != nil {
		return nil, errors.New(fmt.Sprintf("request error: %v", er))
	}

	defer res.Body.Close()
	body, rer := ioutil.ReadAll(res.Body)
	if rer != nil {
		return nil, errors.New(fmt.Sprintf("read body error: %v", rer))
	}

	resp := new(JueJinResponse)
	if er := json.Unmarshal(body, resp); er == nil {
		if resp.Status != "ok" {
			return nil, errors.New("juejin upload error: " + string(body))
		}
		return &Result{
			Url: "https://" + resp.Data.Domain + "/" + resp.Data.Url,
			Raw: body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("decode body error: %v: %s", er, string(body)))
	}
}

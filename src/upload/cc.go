package upload

import (
	"bytes"
	"dxkite.cn/go-storage/src/common"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type CcItem struct {
	Url string `json:"url"`
}

type CcResponse struct {
	Error        int      `json:"total_error"`
	SuccessImage []CcItem `json:"success_image"`
}

const CC = "cc"

type CcUploader struct {
}

func init() {
	Register(CC, func(config Config) Uploader {
		return &CcUploader{}
	})
}

func (*CcUploader) Upload(object *FileObject) (*Result, error) {
	url := "https://upload.cc/image_upload"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if fw, e := w.CreateFormFile("uploaded_file[]", object.Name); e == nil && fw != nil {
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

	//fmt.Println(string(body))

	resp := new(CcResponse)
	if er := json.Unmarshal(body, resp); er == nil {
		if resp.Error != 0 {
			return nil, errors.New("cc upload error: " + string(body))
		}
		return &Result{
			Url: "https://upload.cc/" + resp.SuccessImage[0].Url,
			Raw: body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("decode body error: %v: %s", er, string(body)))
	}
}

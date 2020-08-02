package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
)

const OppoFeedback = "oppo-feedback"

type OppoFeedbackUploader struct {
}

type OppoFeedbackData struct {
	Url string `json:"url"`
}

type OppoFeedbackResponse struct {
	Errno int              `json:"errno"`
	Data  OppoFeedbackData `json:"data"`
}

func init() {
	Register(OppoFeedback, func(config Config) Uploader {
		return &OppoFeedbackUploader{}
	})
}

func (*OppoFeedbackUploader) Upload(object *FileObject) (*Result, error) {
	url := "https://api.open.oppomobile.com/api/utility/upload"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if fw, e := w.CreateFormField("type"); e == nil && fw != nil {
		if _, er := fw.Write([]byte("feedback")); er != nil {
			return nil, er
		}
	}
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

	defer func() { _ = res.Body.Close() }()
	body, rer := ioutil.ReadAll(res.Body)

	if rer != nil {
		return nil, errors.New(fmt.Sprintf("read body error: %v", rer))
	}

	if res.StatusCode != 200 {
		return nil, errors.New("vim-cn status error:" + string(body))
	}

	resp := new(OppoFeedbackResponse)
	if er := json.Unmarshal(body, resp); er == nil {
		if resp.Errno != 0 {
			return nil, errors.New("OppoFeedbackUploader upload error: " + strconv.Itoa(resp.Errno))
		}
		return &Result{
			Url: resp.Data.Url,
			Raw: body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("decode body error: %v: %s", er, string(body)))
	}
}

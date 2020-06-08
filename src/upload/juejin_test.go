package upload

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestJuejinUploader_Upload(t *testing.T) {
	if data, err := ioutil.ReadFile("./test/1.png"); err == nil {
		res, er := Upload(JUEJIN, &FileObject{
			Name: "cdn.png",
			Data: data,
		})
		if er != nil {
			t.Error(er)
		} else {
			r, er := http.Get(res.Url)
			if er != nil {
				t.Error(er)
			}
			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, r.Body); err != nil {
				t.Error(err)
			}
			if bytes.Compare(data, buf.Bytes()) != 0 {
				fmt.Println("uploaded but not raw data", res.Url)
			}
			fmt.Println("uploaded", res.Url)
		}
	} else {
		t.Error(err)
	}
}

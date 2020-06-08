package upload

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestAli_Upload(t *testing.T) {
	if data, err := ioutil.ReadFile("./test/1.png"); err == nil {
		res, er := Upload(ALI, &FileObject{
			Name: "cdn.png",
			Data: data,
		})
		if er != nil {
			t.Error(er)
		} else {
			fmt.Println("uploaded", res.Url)
		}
	} else {
		t.Error(err)
	}
}

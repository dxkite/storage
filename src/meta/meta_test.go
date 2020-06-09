package meta

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestMeta(t *testing.T) {

	info := Info{
		Status:    0,
		Hash:      []byte{1, 2, 3, 4},
		Name:      "dxkite",
		Size:      0,
		BlockSize: 0,
		Encode:    0,
		Type:      0,
		Block: []DataBlock{
			{
				Hash:  []byte{1, 2, 3, 4},
				Index: 1,
				Data:  []byte{5, 6, 7, 8},
			},
			{
				Hash:  []byte{9, 10, 11, 12},
				Index: 2,
				Data:  []byte{13, 14, 15, 16},
			},
		},
	}

	info2 := Info{
		Status:    0,
		Hash:      []byte{1, 2, 3, 4},
		Name:      "dxkite",
		Size:      0,
		BlockSize: 0,
		Encode:    0,
		Type:      0,
		Block: []DataBlock{
			{
				Hash:  []byte{1, 2, 3, 4},
				Index: 1,
				Data:  []byte{5, 6, 7, 8},
			},
			{
				Hash:  []byte{9, 10, 11, 12},
				Index: 2,
				Data:  []byte{13, 14, 15, 16},
			},
		},
	}

	var buf = &bytes.Buffer{}
	if er := EncodeToStream(buf, &info); er != nil {
		t.Error(er)
	}

	if n, er := DecodeFromStream(buf); er != nil {
		t.Error(er)
	} else {
		if reflect.DeepEqual(n, &info2) == false {
			fmt.Printf("want %v\ngot  %v\n", &info2, n)
			t.Error("encode not equal decode")
		}
	}
}

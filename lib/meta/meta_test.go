package meta

import (
	"reflect"
	"testing"
)

func TestEnDecodeToFile(t *testing.T) {
	b := &MetaInfo{
		Hash:   []byte{0, 1, 2, 3, 4, 5},
		Size:   1020,
		Encode: 0,
		Blocks: []DataBlock{
			{
				Hash:  []byte{1, 2, 3, 5, 6, 7, 8, 9, 0},
				Type:  0,
				Index: 1,
				Data:  []byte("hello world"),
			},
		},
	}

	err := EncodeToFile("test/b.meta", b)
	if err != nil {
		t.Error(err)
	}
	if d, der := DecodeToFile("test/b.meta"); der != nil {
		t.Error(der)
	} else {
		if reflect.DeepEqual(b, d) == false {
			t.Errorf("want %v got %v\n", b, d)
		}
	}
}

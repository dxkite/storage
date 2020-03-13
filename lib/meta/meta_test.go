package meta

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEnDecodeToFile(t *testing.T) {
	b := &MetaInfo{
		Hash:   []byte{0, 1, 2, 3, 4, 5},
		Size:   1020,
		Encode: 0,
		Block: []DataBlock{
			{
				Hash:  []byte{1, 2, 3, 5, 6, 7, 8, 9, 0},
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

func TestDecodeToFile(t *testing.T) {
	if d, der := DecodeToFile("test/a.meta"); der != nil {
		t.Error(der)
	} else {
		fmt.Println("file name:", d.Name)
		fmt.Println("file size:", d.Size)
		fmt.Printf("file hash: %x\n", d.Hash)
		fmt.Println("file status:", d.Status)
		for _, b := range d.Block {
			fmt.Printf("block %d hash %x\n", b.Index, b.Hash)
			fmt.Printf("block %d link %s\n", b.Index, string(b.Data))
		}
	}
}

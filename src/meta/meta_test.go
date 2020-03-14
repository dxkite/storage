package meta

import (
	"dxkite.cn/go-storage/storage"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

func TestDecodeToFile(t *testing.T) {
	if d, der := DecodeToFile("test/a.meta"); der != nil {
		t.Error(der)
	} else {
		h1, _ := hex.DecodeString("0d14a3da07c74efaee62f3ea495ce7de2e62c257")
		h2, _ := hex.DecodeString("ee8e89cf59ac0fcc22639e6d24095bb5f1655e64")
		h3, _ := hex.DecodeString("06e1c1d7d54df2d78f5a5214a5d063422b3a798b")

		r := &MetaInfo{
			Status:    Finish,
			Hash:      h1,
			Name:      "陈奕迅 - 十年.mp3",
			Size:      3242822,
			BlockSize: 2097152,
			Type:      int32(storage.DataResponse_URI),
			Encode:    int32(storage.DataResponse_IMAGE),
			Block: []DataBlock{
				{
					Index: 0,
					Hash:  h2,
					Data:  []byte("https://ae01.alicdn.com/kf/U970f381e47dd4ef3a502f11ad574ee4ei.png"),
				},
				{
					Index: 1,
					Hash:  h3,
					Data:  []byte("https://ae01.alicdn.com/kf/U6f5fedea92c0464a846f72a420050280t.png"),
				},
			},
		}
		fmt.Println("file name:", d.Name)
		fmt.Println("file size:", d.Size)
		fmt.Println("file block size:", d.BlockSize)
		fmt.Printf("file hash: %x\n", d.Hash)
		fmt.Println("file status:", d.Status)
		for _, b := range d.Block {
			fmt.Printf("block %d hash %x\n", b.Index, b.Hash)
			fmt.Printf("block %d link %s\n", b.Index, string(b.Data))
		}
		if reflect.DeepEqual(r, d) == false {
			t.Errorf("want %v got %v\n", r, d)
		}
	}
}

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

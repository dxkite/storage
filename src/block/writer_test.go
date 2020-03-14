package block

import (
	"crypto/sha1"
	"os"
	"testing"
)

func TestBlockFile_WriteRange(t *testing.T) {
	var r1 = []byte("hello world")
	var r2 = []byte(", this is the test")
	var r3 = []byte(" range File")

	var data = r1
	data = append(data, r2...)
	data = append(data, r3...)

	h := sha1.New()
	h.Write(data)
	var hash = h.Sum(nil)

	file, _ := os.OpenFile("test/test.txt", os.O_CREATE|os.O_RDWR, os.ModePerm)

	var block = BlockFile{
		Hash: hash,
		File: file,
	}

	if err := block.WriteBlock(NewBlock(int64(len(r1)), int64(len(r1)+len(r2)), r2)); err != nil {
		t.Error(err)
	}
	if err := block.WriteBlock(NewBlock(0, int64(len(r1)), r1)); err != nil {
		t.Error(err)
	}
	if err := block.WriteBlock(NewBlock(int64(len(r1)+len(r2)), int64(len(r1)+len(r2)+len(r3)), r3)); err != nil {
		t.Error(err)
	}

	if block.CheckSum() == false {
		t.Error("block write error")
	}
}

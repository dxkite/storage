package image

import (
	"bytes"
	"testing"
)

func TestEnDecodeByte(t *testing.T) {
	raw := []byte("hello world")
	b, eer := EncodeByte(raw)
	if eer != nil {
		t.Error(eer)
	}
	d, der := DecodeByte(b)
	if der != nil {
		t.Error(der)
	}
	if bytes.Equal(raw, d) == false {
		t.Errorf("want %v go %v\n", raw, d)
	}
}

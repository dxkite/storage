package image

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeByte(t *testing.T) {
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

func TestEncodeSize(t *testing.T) {
	raw := []byte("hello world, long, long.")
	b, eer := EncodeByteSize(raw, 16)
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

func TestEncodeSize_64(t *testing.T) {
	raw := []byte(strings.Repeat("hello world", 64))
	b, eer := EncodeByteSize(raw, 16)
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

func TestEncodeSize_512(t *testing.T) {
	raw := []byte(strings.Repeat("hello world", 512))
	b, eer := EncodeByteSize(raw, 16)
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

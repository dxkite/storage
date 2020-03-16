package meta

import "io"

type XORWrapper struct {
	io.ReadWriter
	Value byte
}

func NewXor(n byte, rw io.ReadWriter) io.ReadWriter {
	return XORWrapper{
		ReadWriter: rw,
		Value:      n,
	}
}

// 写包装
func (c XORWrapper) Read(b []byte) (n int, err error) {
	n, re := c.ReadWriter.Read(b)
	if re != nil {
		err = re
		return
	}
	for i, v := range b {
		b[i] = v ^ c.Value
	}
	return n, err
}

// 读包装
func (c XORWrapper) Write(b []byte) (n int, err error) {
	for i, v := range b {
		b[i] = v ^ c.Value
	}
	return c.ReadWriter.Write(b)
}

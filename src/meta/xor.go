package meta

import "io"

type XORReader struct {
	io.Reader
	Value byte
}

func NewXORReader(n byte, rw io.Reader) io.Reader {
	return XORReader{
		Reader: rw,
		Value:  n,
	}
}

// 写包装
func (c XORReader) Read(b []byte) (n int, err error) {
	n, re := c.Reader.Read(b)
	if re != nil {
		err = re
		return
	}
	for i, v := range b {
		b[i] = v ^ c.Value
	}
	return n, err
}

type XORWriter struct {
	io.Writer
	Value byte
}

func NewXORWriter(n byte, rw io.ReadWriter) io.Writer {
	return XORWriter{
		Writer: rw,
		Value:  n,
	}
}

// 读包装
func (c XORWriter) Write(b []byte) (n int, err error) {
	for i, v := range b {
		b[i] = v ^ c.Value
	}
	return c.Writer.Write(b)
}

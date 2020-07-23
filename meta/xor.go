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

func (c XORReader) Read(b []byte) (n int, err error) {
	n, err = c.Reader.Read(b)
	if err != nil {
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

func NewXORWriter(n byte, w io.Writer) io.Writer {
	return XORWriter{
		Writer: w,
		Value:  n,
	}
}

func (c XORWriter) Write(b []byte) (n int, err error) {
	for i, v := range b {
		b[i] = v ^ c.Value
	}
	return c.Writer.Write(b)
}

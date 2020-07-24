package image

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
)

type imageReadWriter struct {
	Image   *image.NRGBA
	Version uint32
	Length  uint32
	offset  int
}

const (
	HeadVersion   = 4
	ContentLength = 4
	HeaderLength  = HeadVersion + ContentLength
)

type Writer struct {
	Image   *image.NRGBA
	Version uint32
	Length  uint32
	offset  int
}

func (p *imageReadWriter) Write(b []byte) (n int, err error) {
	r := false
	of := HeaderLength + p.offset
	s := len(p.Image.Pix) - of
	lb := len(b)
	if lb > s {
		r = true
		// 复制有的内容
		copy(p.Image.Pix[of:], b[:s])
		// 处理剩余空间
		p.Image.Pix = append(p.Image.Pix, b[s:]...)
	} else {
		copy(p.Image.Pix[of:], b[:])
	}
	p.offset += lb
	p.Length += uint32(lb)
	// 重置大小
	if r {
		p.resize()
	}
	p.writeHeader()
	return lb, nil
}

func (p *imageReadWriter) writeHeader() {
	if len(p.Image.Pix) < HeaderLength {
		p.Image.Pix = make([]byte, HeaderLength)
	}
	binary.BigEndian.PutUint32(p.Image.Pix[0:4:4], p.Version)
	binary.BigEndian.PutUint32(p.Image.Pix[4:8:8], p.Length)
}

func (p *imageReadWriter) readHeader() (version, length uint32) {
	version = binary.BigEndian.Uint32(p.Image.Pix[0:4:4])
	length = binary.BigEndian.Uint32(p.Image.Pix[4:8:8])
	return
}

func (p *imageReadWriter) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	p.Version, p.Length = p.readHeader()
	of := HeaderLength + p.offset
	s := int(p.Length) - p.offset
	if s == 0 {
		return 0, io.EOF
	}
	lb := len(b)
	if s > lb {
		n = copy(b[:], p.Image.Pix[of:of+lb])
	} else {
		n = copy(b[:s], p.Image.Pix[of:])
	}
	p.offset += n
	return n, nil
}

func NewEncoder(size int) *imageReadWriter {
	w, h := getSize(size)
	img := &imageReadWriter{
		Image:   image.NewNRGBA(image.Rect(0, 0, w, h)),
		offset:  0,
		Version: 1,
		Length:  0,
	}
	img.writeHeader()
	return img
}

func (p *imageReadWriter) resize() *image.NRGBA {
	w, h := getSize(HeaderLength + int(p.Length))
	p.Image.Stride = 4 * w
	p.Image.Rect = image.Rect(0, 0, w, h)
	l := w * h * 4
	pl := len(p.Image.Pix)
	if pl > l {
		p.Image.Pix = p.Image.Pix[0:l]
	} else {
		e := make([]byte, l-pl)
		p.Image.Pix = append(p.Image.Pix, e...)
	}
	return p.Image
}

func getSize(size int) (w, h int) {
	x := math.Ceil(float64(size) / 4)
	d := math.Sqrt(float64(x))
	return int(math.Ceil(d)), int(math.Ceil(d))
}

// 解码
func Decode(r io.Reader, w io.Writer) error {
	i, err := png.Decode(r)
	if err != nil {
		return err
	}
	switch i := i.(type) {
	case *image.NRGBA:
		p := &imageReadWriter{Image: i}
		if _, err := io.Copy(w, p); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("unknown color mode：image.RGBA"))
	}
	return nil
}

// 编码
func Encode(w io.Writer, r io.Reader) error {
	return EncodeSize(w, r, 1024)
}

// 直接编码
func EncodeSize(w io.Writer, r io.Reader, size int) error {
	p := NewEncoder(size)
	if _, err := io.Copy(p, r); err != nil {
		return err
	}
	p.resize()
	return png.Encode(w, p.Image)
}

// 编码
func EncodeByte(input []byte) ([]byte, error) {
	return EncodeByteSize(input, 1024)
}

// 编码
func EncodeByteSize(input []byte, size int) ([]byte, error) {
	i := bytes.NewBuffer(input)
	o := &bytes.Buffer{}
	err := EncodeSize(o, i, size)
	if err != nil {
		return nil, err
	}
	return o.Bytes(), nil
}

// 编码
func DecodeByte(input []byte) ([]byte, error) {
	i := bytes.NewBuffer(input)
	o := &bytes.Buffer{}
	err := Decode(i, o)
	if err != nil {
		return nil, err
	}
	return o.Bytes(), nil
}

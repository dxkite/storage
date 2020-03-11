package encoding

import (
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
	n = p.offset
	r := false
	h := HeaderLength
	// 写入数据
	for _, d := range b {
		if p.offset+h < len(p.Image.Pix) {
			p.Image.Pix[h+p.offset] = d
		} else {
			p.Image.Pix = append(p.Image.Pix, d)
			r = true
		}
		p.offset++
	}
	// 重置大小
	if r {
		p.resize()
	}
	d := p.offset - n
	p.Length += uint32(d)
	p.writeHeader()
	return d, nil
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
	n = p.offset
	h := HeaderLength
	p.Version, p.Length = p.readHeader()
	for i, _ := range b {
		if uint32(p.offset) < p.Length {
			b[i] = p.Image.Pix[h+p.offset]
			p.offset++
		} else {
			break
		}
	}
	if p.offset-n == 0 {
		return 0, io.EOF
	}
	return p.offset - n, nil
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
	return p.Image
}

func getSize(size int) (w, h int) {
	x := size / 4
	d := math.Sqrt(float64(x))
	return int(math.Ceil(d)), int(math.Ceil(d))
}

type Encoder struct {
}

// 解码
func (*Encoder) Decode(r io.Reader, w io.Writer) error {
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
		return errors.New(fmt.Sprintf("unknown color mode： image.RGBA"))
	}
	return nil
}

// 编码
func (*Encoder) Encode(w io.Writer, r io.Reader) error {
	p := NewEncoder(1024)
	if _, err := io.Copy(p, r); err != nil {
		return err
	}
	return png.Encode(w, p.resize())
}

var StdImageEncoding = &Encoder{}

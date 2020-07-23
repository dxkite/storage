package bitset

import "math"

type BitSet []byte

func New(size int64) BitSet {
	return make([]byte, int(math.Ceil(float64(size)/8)))
}

func (b BitSet) Get(index int64) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= int64(len(b)) {
		return false
	}
	return b[byteIndex]>>(7-offset)&1 != 0
}

func (b BitSet) Set(index int64) {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= int64(len(b)) {
		return
	}
	b[byteIndex] |= 1 << (7 - offset)
}

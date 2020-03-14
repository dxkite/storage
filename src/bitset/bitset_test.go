package bitset

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetBit(t *testing.T) {
	set := BitSet{0b01010100, 0b01010100}
	outputs := []bool{false, true, false, true, false, true, false, false, false, true, false, true, false, true, false, false, false, false, false, false}
	for i := 0; i < len(outputs); i++ {
		if outputs[i] != set.Get(int64(i)) {
			t.Errorf("want %v got %v", outputs[i], set.Get(int64(i)))
		}
	}
}

func TestSetBit(t *testing.T) {
	tests := []struct {
		input  BitSet
		index  int64
		output BitSet
	}{
		{
			input:  BitSet{0b01010100, 0b01010100},
			index:  4,
			output: BitSet{0b01011100, 0b01010100},
		},
		{
			input:  BitSet{0b01010100, 0b01010100},
			index:  9,
			output: BitSet{0b01010100, 0b01010100},
		},
		{
			input:  BitSet{0b01010100, 0b01010100},
			index:  15,
			output: BitSet{0b01010100, 0b01010101},
		},
		{
			input:  BitSet{0b01010100, 0b01010100},
			index:  19,
			output: BitSet{0b01010100, 0b01010100},
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("bit bitset #%d", i), func(t *testing.T) {
			r := test.input
			r.Set(test.index)
			if reflect.DeepEqual(r, test.output) == false {
				t.Errorf("want %v got %v", test.output, r)
			}
		})
	}
}

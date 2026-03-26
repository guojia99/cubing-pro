package utils_tool

import "math/bits"

type Bitset struct {
	data []uint64
}

func NewBitset(size int) *Bitset {
	return &Bitset{
		data: make([]uint64, (size+63)>>6),
	}
}

func (b *Bitset) Set(i int) {
	b.data[i>>6] |= 1 << (i & 63)
}

func (b *Bitset) And(other *Bitset) *Bitset {
	res := &Bitset{data: make([]uint64, len(b.data))}
	for i := range b.data {
		res.data[i] = b.data[i] & other.data[i]
	}
	return res
}

func (b *Bitset) ForEach(fn func(i int)) {
	for wordIdx, word := range b.data {
		for word != 0 {
			t := word & -word
			bit := bits.TrailingZeros64(word)
			idx := wordIdx*64 + bit
			fn(idx)
			word ^= t
		}
	}
}

// Or 返回两个 Bitset 的按位或结果，长度必须相同
func (b *Bitset) Or(other *Bitset) *Bitset {
	if len(b.data) != len(other.data) {
		panic("Bitset length mismatch in Or")
	}

	result := &Bitset{
		data: make([]uint64, len(b.data)),
	}

	for i := range b.data {
		result.data[i] = b.data[i] | other.data[i]
	}

	return result
}

package main

import "math/bits"

type bitMask uint64

func (b *bitMask) set(index int) {
	*b |= (1 << index)
}

func (b *bitMask) unset(index int) {
	*b &= ^(1 << index)
}

func (b bitMask) has(index int) bool {
	return (b & (1 << index)) != 0
}

func (b *bitMask) toggle(index int) {
	if b.has(index) {
		b.unset(index)
	} else {
		b.set(index)
	}
}

func (b *bitMask) clear() {
	*b = 0
}

func (b bitMask) isEmpty() bool {
	return b == 0
}

func (b bitMask) count() int {
	return bits.OnesCount64(uint64(b))
}

func (b bitMask) indexes() []int {
	var indexes []int

	for b != 0 {
		i := bits.TrailingZeros64(uint64(b))
		indexes = append(indexes, i)
		b &^= 1 << i
	}

	return indexes
}

func (b bitMask) contains(other bitMask) bool {
	return b&other == other
}

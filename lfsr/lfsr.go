// Package lfsr defines the abstractions shared by the linear feedback shift
// register implementations in this module.
//
// A linear feedback shift register is a shift register whose input bit is a
// linear function of its previous state, which makes it a compact source of
// pseudo-random bits, pseudo-noise sequences and whitening sequences.
package lfsr

import "math/bits"

// Word enumerates the register widths an LFSR can be built on. The width of W
// fixes the degree of the feedback polynomial, and therefore the longest period
// the register can reach, 2^n - 1.
type Word interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Width reports the number of bits in W, which is the degree of the feedback
// polynomial a register of that width runs.
func Width[W Word]() int {
	return bits.Len64(uint64(^W(0)))
}

// ReverseMask converts a tap mask between the Fibonacci and Galois conventions.
//
// The two forms number their taps from opposite ends: a Fibonacci mask has bit i
// set for the term x^(n-1-i), while a Galois mask has bit j set for the term
// x^j. For one and the same polynomial the two masks are therefore bit
// reversals of each other over the width of W, and ReverseMask is its own
// inverse.
//
// The Fibonacci mask 0xB8 and the Galois mask 0x1D both run
// x^8 + x^4 + x^3 + x^2 + 1; ReverseMask maps either onto the other.
func ReverseMask[W Word](mask W) W {
	return W(bits.Reverse64(uint64(mask)) >> (64 - Width[W]()))
}

// LFSR is a linear feedback shift register over a W-wide state.
//
// Implementations are not safe for concurrent use.
type LFSR[W Word] interface {
	// Next clocks the register once and reports the bit shifted out, which is
	// always 0 or 1. This is the register's pseudo-random output.
	Next() uint8

	// Uint clocks the register once per bit of W and packs the outgoing bits
	// into a word, most significant bit first. Successive words are drawn from
	// disjoint runs of the bit stream, so they do not overlap the way raw
	// register states do.
	Uint() W

	// State reports the current register contents without clocking it.
	State() W
}

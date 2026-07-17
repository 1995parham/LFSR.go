// Package flfsr implements a Fibonacci linear feedback shift register over any
// unsigned word width.
package flfsr

import (
	"errors"
	"math/bits"

	"github.com/1995parham/LFSR.go/lfsr"
)

var (
	// ErrZeroPoly is returned when a register is built with no taps. Such a
	// register only shifts zeros in and dies at the all-zero state.
	ErrZeroPoly = errors.New("flfsr: polynomial must be non-zero")

	// ErrZeroSeed is returned when a register is seeded with zero. The all-zero
	// state is a fixed point: every subsequent output would be zero.
	ErrZeroSeed = errors.New("flfsr: seed must be non-zero")
)

// FLFSR is a Fibonacci linear feedback shift register over a W-wide state.
//
// The register shifts towards the most significant bit. Each clock emits the
// outgoing top bit and feeds the parity of the tapped bits back into bit 0.
//
// The zero value is not usable; build one with [New].
type FLFSR[W lfsr.Word] struct {
	data W
	poly W
}

var _ lfsr.LFSR[uint8] = (*FLFSR[uint8])(nil)

// New builds a Fibonacci LFSR of the width of W.
//
// poly is a tap mask over register bit positions, not a table of polynomial
// exponents. Because the register shifts up, mask bit i stands for the term
// x^(n-1-i) of the feedback polynomial, where n is the width of W. The x^n term
// is implicit, being the feedback itself, while the polynomial's constant term
// lands on bit n-1; a primitive polynomial always has a constant term, so the
// top bit of a usable mask is always set. Mask 0xB8 taps register bits 7, 5, 4
// and 3, which is the degree-8 polynomial x^8 + x^4 + x^3 + x^2 + 1.
//
// Most published tap tables describe right-shifting registers. Such a mask used
// here yields the reciprocal of the polynomial the table names. That costs
// nothing in practice, since the reciprocal of a primitive polynomial is itself
// primitive and has the same period, but it does mean the sequence is not the
// one the table describes.
//
// Reaching the maximal period of 2^n - 1 requires poly to be primitive over
// GF(2); New only rejects the degenerate masks, it does not check primitivity.
// Any non-zero seed will do, and the seed is itself a state on the cycle.
func New[W lfsr.Word](poly, seed W) (*FLFSR[W], error) {
	if poly == 0 {
		return nil, ErrZeroPoly
	}

	if seed == 0 {
		return nil, ErrZeroSeed
	}

	return &FLFSR[W]{data: seed, poly: poly}, nil
}

// Next clocks the register once and reports the bit shifted out of the top.
func (f *FLFSR[W]) Next() uint8 {
	out := uint8(f.data>>(width[W]()-1)) & 1

	// A Fibonacci LFSR feeds back the XOR of every tapped bit, which is the
	// parity of the tapped bits taken together.
	feedback := W(bits.OnesCount64(uint64(f.data&f.poly)) & 1)
	f.data = f.data<<1 | feedback

	return out
}

// Uint clocks the register once per bit of W and packs the outgoing bits into a
// word, most significant bit first.
func (f *FLFSR[W]) Uint() W {
	var v W
	for range width[W]() {
		v = v<<1 | W(f.Next())
	}

	return v
}

// State reports the current register contents without clocking it.
func (f *FLFSR[W]) State() W {
	return f.data
}

// width reports the number of bits in W.
func width[W lfsr.Word]() int {
	return bits.Len64(uint64(^W(0)))
}

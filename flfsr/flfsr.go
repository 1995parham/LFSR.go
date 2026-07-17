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
// exponents: setting bit i taps register bit i, which stands for the term
// x^(i+1) of the feedback polynomial. The implicit +1 term of the polynomial is
// the incoming feedback bit and is never expressed in the mask. So the degree-8
// polynomial x^8 + x^6 + x^5 + x^4 + 1 taps register bits 7, 5, 4 and 3, giving
// poly 0xB8.
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

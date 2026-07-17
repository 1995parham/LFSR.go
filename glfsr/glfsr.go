// Package glfsr implements a Galois linear feedback shift register over any
// unsigned word width.
//
// The Galois form is the counterpart of the Fibonacci form in
// [github.com/1995parham/LFSR.go/flfsr]. Driven by the same polynomial the two
// run the same sequence, only at a different phase, so the choice between them
// is a matter of cost rather than quality: a Fibonacci clock folds every tapped
// bit into one feedback bit, while a Galois clock shifts and, when the outgoing
// bit is set, XORs the whole mask into the state at once.
package glfsr

import (
	"errors"

	"github.com/1995parham/LFSR.go/lfsr"
)

var (
	// ErrZeroPoly is returned when a register is built with no taps. Such a
	// register only shifts zeros in and dies at the all-zero state.
	ErrZeroPoly = errors.New("glfsr: polynomial must be non-zero")

	// ErrZeroSeed is returned when a register is seeded with zero. The all-zero
	// state is a fixed point: every subsequent output would be zero.
	ErrZeroSeed = errors.New("glfsr: seed must be non-zero")
)

// GLFSR is a Galois linear feedback shift register over a W-wide state.
//
// The register shifts towards the most significant bit. Each clock emits the
// outgoing top bit and, when that bit is set, XORs the tap mask into the state.
//
// The zero value is not usable; build one with [New].
type GLFSR[W lfsr.Word] struct {
	data W
	poly W
}

var _ lfsr.LFSR[uint8] = (*GLFSR[uint8])(nil)

// New builds a Galois LFSR of the width of W.
//
// poly is a tap mask holding the low n coefficients of the feedback polynomial,
// where n is the width of W: bit j stands for the term x^j. The x^n term is
// implicit, being the outgoing bit that triggers the XOR. So the degree-8
// polynomial x^8 + x^4 + x^3 + x^2 + 1 gives poly 0x1D. A primitive polynomial
// always has a constant term, so bit 0 of a usable mask is always set.
//
// Note that this is not the convention [github.com/1995parham/LFSR.go/flfsr]
// uses. The two forms number their taps from opposite ends, so the same
// polynomial is 0x1D here and 0xB8 there. Feeding one form the other's mask
// still runs, but it runs a different, most likely non-primitive polynomial and
// so gives a far shorter period. Use [lfsr.ReverseMask] to convert.
//
// Reaching the maximal period of 2^n - 1 requires poly to be primitive over
// GF(2); New only rejects the degenerate masks, it does not check primitivity.
// Any non-zero seed will do, and the seed is itself a state on the cycle.
func New[W lfsr.Word](poly, seed W) (*GLFSR[W], error) {
	if poly == 0 {
		return nil, ErrZeroPoly
	}

	if seed == 0 {
		return nil, ErrZeroSeed
	}

	return &GLFSR[W]{data: seed, poly: poly}, nil
}

// Next clocks the register once and reports the bit shifted out of the top.
func (g *GLFSR[W]) Next() uint8 {
	out := uint8(g.data>>(lfsr.Width[W]()-1)) & 1

	// Shifting the outgoing bit out is what feeds it back: every tapped term of
	// the polynomial flips at once, rather than being folded into a single
	// incoming bit as the Fibonacci form does.
	//
	// Negating the outgoing bit gives all ones when it was set and zero when it
	// was not, which applies the mask exactly when it is due. The obvious "if
	// out == 1" spelling is a coin-flip branch that no predictor can learn, and
	// it measures around 2.5x slower than this.
	g.data = g.data<<1 ^ (g.poly & -W(out))

	return out
}

// Uint clocks the register once per bit of W and packs the outgoing bits into a
// word, most significant bit first.
func (g *GLFSR[W]) Uint() W {
	var v W
	for range lfsr.Width[W]() {
		v = v<<1 | W(g.Next())
	}

	return v
}

// State reports the current register contents without clocking it.
func (g *GLFSR[W]) State() W {
	return g.data
}

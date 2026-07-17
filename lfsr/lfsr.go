// Package lfsr defines the abstractions shared by the linear feedback shift
// register implementations in this module.
//
// A linear feedback shift register is a shift register whose input bit is a
// linear function of its previous state, which makes it a compact source of
// pseudo-random bits, pseudo-noise sequences and whitening sequences.
package lfsr

// Word enumerates the register widths an LFSR can be built on. The width of W
// fixes the degree of the feedback polynomial, and therefore the longest period
// the register can reach, 2^n - 1.
type Word interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
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

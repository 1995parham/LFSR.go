// Package prng builds a pseudo-random bit stream out of a pair of linear
// feedback shift registers.
//
// A lone LFSR is not a pseudo-random generator worth the name. It is linear, and
// that is fatal: Berlekamp-Massey recovers the whole feedback polynomial, and
// with it every future bit, from only 2n output bits. Generators like the one
// here exist to destroy that linearity by combining two registers so that the
// output is no longer a linear function of either state.
//
// This is a port of the prng.c sample from [the C implementation] this module is
// based on, which used to be vendored here. Read [Shrinking] before using it for
// anything that matters.
//
// [the C implementation]: https://github.com/mfukar/lfsr
package prng

import (
	"errors"

	"github.com/1995parham/LFSR.go/lfsr"
)

// ErrNilRegister is returned when a generator is built without both registers.
var ErrNilRegister = errors.New("prng: both the data and the clock register are required")

// Shrinking is a shrinking generator, as described by Coppersmith, Krawczyk and
// Mansour.
//
// Both registers clock together. When the clock register emits a one the data
// register's bit is passed through; when it emits a zero that data bit is thrown
// away. Roughly half the data bits therefore never reach the caller, which is
// what breaks the linearity a single register suffers from, and is also why the
// output rate is irregular.
//
// # This generator is not cryptographically sound as this module can build it
//
// The shrinking generator only reaches its full period of (2^m - 1) * 2^(n-1)
// when the two registers' periods are coprime, which needs gcd(m, n) = 1 for
// register lengths m and n. In this module a register's length is the width of
// its word type, so the only lengths available are 8, 16, 32 and 64. Every pair
// of those shares a factor of at least 8, so no combination is coprime and every
// generator built here falls short of its period, badly. An 8-bit pair repeats
// after about 128 bits rather than 32640.
//
// The C original avoided this by storing an arbitrary-degree polynomial in a
// fixed 64-bit word, which lets it pick coprime lengths such as 17 and 19. Doing
// the same here would mean decoupling the polynomial's degree from the word
// type. Until then, treat this as a demonstration of the construction, not as a
// source of secrets.
type Shrinking[D, C lfsr.Word] struct {
	data  lfsr.LFSR[D]
	clock lfsr.LFSR[C]
}

// NewShrinking builds a shrinking generator over a data and a clock register.
//
// The two need not be the same width, and need not be the same form: mixing an
// [github.com/1995parham/LFSR.go/flfsr] with an
// [github.com/1995parham/LFSR.go/glfsr] is fine.
func NewShrinking[D, C lfsr.Word](data lfsr.LFSR[D], clock lfsr.LFSR[C]) (*Shrinking[D, C], error) {
	if data == nil || clock == nil {
		return nil, ErrNilRegister
	}

	return &Shrinking[D, C]{data: data, clock: clock}, nil
}

// Next clocks both registers until the clock register emits a one, and reports
// the data bit that came out alongside it.
//
// This always terminates. A register only ever shifts its bits towards the top,
// and a non-zero register never reaches the all-zero state, so a one must leave
// the clock register within its own width in clocks.
func (s *Shrinking[D, C]) Next() uint8 {
	for {
		bit := s.data.Next()

		if s.clock.Next() == 1 {
			return bit
		}
	}
}

// Uint clocks the generator once per bit of D and packs the surviving bits into
// a word, most significant bit first.
func (s *Shrinking[D, C]) Uint() D {
	var v D
	for range lfsr.Width[D]() {
		v = v<<1 | D(s.Next())
	}

	return v
}

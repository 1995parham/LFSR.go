package prng_test

import (
	"errors"
	"testing"

	"github.com/1995parham/LFSR.go/flfsr"
	"github.com/1995parham/LFSR.go/glfsr"
	"github.com/1995parham/LFSR.go/lfsr"
	"github.com/1995parham/LFSR.go/prng"
)

const (
	fibPoly8  uint8  = 0xB8   // x^8 + x^4 + x^3 + x^2 + 1
	fibPoly16 uint16 = 0xB400 // x^16 + x^5 + x^3 + x^2 + 1
	galPoly8  uint8  = 0x1D   // the same polynomial as fibPoly8, galois spelling
)

func mustFib8(t *testing.T, seed uint8) lfsr.LFSR[uint8] {
	t.Helper()

	f, err := flfsr.New(fibPoly8, seed)
	if err != nil {
		t.Fatalf("flfsr.New: %v", err)
	}

	return f
}

func TestNewShrinkingRejectsNilRegisters(t *testing.T) {
	t.Parallel()

	if _, err := prng.NewShrinking[uint8, uint8](nil, mustFib8(t, 0x40)); !errors.Is(err, prng.ErrNilRegister) {
		t.Errorf("NewShrinking(nil, clock) error = %v, want %v", err, prng.ErrNilRegister)
	}

	if _, err := prng.NewShrinking[uint8, uint8](mustFib8(t, 0x40), nil); !errors.Is(err, prng.ErrNilRegister) {
		t.Errorf("NewShrinking(data, nil) error = %v, want %v", err, prng.ErrNilRegister)
	}
}

// TestNextMatchesReference pins the generator against the rule the upstream C
// sample states: clock both registers, keep the data bit when the clock bit is
// one and drop it otherwise.
func TestNextMatchesReference(t *testing.T) {
	t.Parallel()

	g, err := prng.NewShrinking(mustFib8(t, 0x40), mustFib8(t, 0xACE1&0xFF))
	if err != nil {
		t.Fatalf("NewShrinking: %v", err)
	}

	refData := mustFib8(t, 0x40)
	refClock := mustFib8(t, 0xACE1&0xFF)

	for i := range 256 {
		got := g.Next()

		var want uint8

		for {
			bit := refData.Next()
			if refClock.Next() == 1 {
				want = bit

				break
			}
		}

		if got != want {
			t.Fatalf("bit %d: Next() = %d, want %d", i, got, want)
		}
	}
}

// TestMixedFormsAndWidths checks the generator composes whatever the module can
// build, rather than only a matched pair.
func TestMixedFormsAndWidths(t *testing.T) {
	t.Parallel()

	data, err := flfsr.New(fibPoly8, uint8(0x40))
	if err != nil {
		t.Fatalf("flfsr.New: %v", err)
	}

	clock, err := glfsr.New(galPoly8, uint8(0x17))
	if err != nil {
		t.Fatalf("glfsr.New: %v", err)
	}

	g, err := prng.NewShrinking[uint8, uint8](data, clock)
	if err != nil {
		t.Fatalf("NewShrinking: %v", err)
	}

	var ones int

	for range 512 {
		if g.Next() == 1 {
			ones++
		}
	}

	if ones == 0 || ones == 512 {
		t.Errorf("stream is constant: %d ones in 512 bits", ones)
	}

	wide, err := flfsr.New(fibPoly16, uint16(0xACE1))
	if err != nil {
		t.Fatalf("flfsr.New: %v", err)
	}

	mixed, err := prng.NewShrinking[uint8, uint16](mustFib8(t, 0x40), wide)
	if err != nil {
		t.Fatalf("NewShrinking: %v", err)
	}

	mixed.Uint()
}

// TestShortPeriodFromSharedFactors is the uncomfortable one. It documents, and
// pins, the weakness described on prng.Shrinking: because this module ties a
// register's length to its word width, two registers of the same width have the
// same period, the periods are not coprime, and the generator collapses to a
// fraction of the (2^m - 1) * 2^(n-1) it is supposed to reach.
//
// If this test ever fails because the period got longer, the module has grown a
// way to pick coprime lengths and the documentation needs revisiting.
func TestShortPeriodFromSharedFactors(t *testing.T) {
	t.Parallel()

	g, err := prng.NewShrinking(mustFib8(t, 0x40), mustFib8(t, 0x17))
	if err != nil {
		t.Fatalf("NewShrinking: %v", err)
	}

	const sample = 4096

	stream := make([]uint8, sample)
	for i := range stream {
		stream[i] = g.Next()
	}

	got := repeatLength(stream)

	const ideal = (1<<8 - 1) * (1 << 7) // 32640, what a coprime pair would give

	if got >= ideal {
		t.Fatalf("period %d reached the coprime ideal %d; the width/degree coupling "+
			"described on prng.Shrinking no longer holds", got, ideal)
	}

	t.Logf("output period = %d bits; a coprime pair would give %d (%.0fx longer)",
		got, ideal, float64(ideal)/float64(got))
}

// repeatLength reports the smallest p for which the stream repeats with period
// p over the whole sample, or len(stream) if it never does.
func repeatLength(stream []uint8) int {
	for p := 1; p < len(stream)/2; p++ {
		ok := true

		for i := range len(stream) - p {
			if stream[i] != stream[i+p] {
				ok = false

				break
			}
		}

		if ok {
			return p
		}
	}

	return len(stream)
}

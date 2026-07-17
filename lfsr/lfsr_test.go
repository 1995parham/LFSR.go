package lfsr_test

import (
	"testing"

	"github.com/1995parham/LFSR.go/lfsr"
)

func TestWidth(t *testing.T) {
	t.Parallel()

	if got, want := lfsr.Width[uint8](), 8; got != want {
		t.Errorf("Width[uint8]() = %d, want %d", got, want)
	}

	if got, want := lfsr.Width[uint16](), 16; got != want {
		t.Errorf("Width[uint16]() = %d, want %d", got, want)
	}

	if got, want := lfsr.Width[uint32](), 32; got != want {
		t.Errorf("Width[uint32]() = %d, want %d", got, want)
	}

	if got, want := lfsr.Width[uint64](), 64; got != want {
		t.Errorf("Width[uint64]() = %d, want %d", got, want)
	}
}

// TestReverseMaskKnownPairs pins the conversion against the mask pairs the two
// packages document, which are the ones a reader is most likely to check.
func TestReverseMaskKnownPairs(t *testing.T) {
	t.Parallel()

	// x^8 + x^4 + x^3 + x^2 + 1
	if got, want := lfsr.ReverseMask[uint8](0xB8), uint8(0x1D); got != want {
		t.Errorf("ReverseMask[uint8](0xB8) = %#x, want %#x", got, want)
	}

	// x^16 + x^5 + x^3 + x^2 + 1
	if got, want := lfsr.ReverseMask[uint16](0xB400), uint16(0x2D); got != want {
		t.Errorf("ReverseMask[uint16](0xB400) = %#x, want %#x", got, want)
	}
}

// TestReverseMaskIsAnInvolution checks the property the doc comment claims, at
// every width: reversing twice must land back where it started.
func TestReverseMaskIsAnInvolution(t *testing.T) {
	t.Parallel()

	t.Run("8-bit", func(t *testing.T) {
		t.Parallel()

		// Exhaustive at this width; there are only 256 masks.
		for m := range 1 << 8 {
			if got := lfsr.ReverseMask(lfsr.ReverseMask(uint8(m))); got != uint8(m) {
				t.Fatalf("ReverseMask twice on %#x = %#x", m, got)
			}
		}
	})

	t.Run("wider", func(t *testing.T) {
		t.Parallel()

		for _, m := range []uint16{0x0001, 0x8000, 0xB400, 0xACE1, 0xFFFF} {
			if got := lfsr.ReverseMask(lfsr.ReverseMask(m)); got != m {
				t.Errorf("ReverseMask twice on %#x = %#x", m, got)
			}
		}

		for _, m := range []uint32{0x00000001, 0x80000000, 0xE0000200, 0xFFFFFFFF} {
			if got := lfsr.ReverseMask(lfsr.ReverseMask(m)); got != m {
				t.Errorf("ReverseMask twice on %#x = %#x", m, got)
			}
		}

		for _, m := range []uint64{0x1, 0x8000000000000000, 0x800000000000000D, ^uint64(0)} {
			if got := lfsr.ReverseMask(lfsr.ReverseMask(m)); got != m {
				t.Errorf("ReverseMask twice on %#x = %#x", m, got)
			}
		}
	})
}

// TestReverseMaskStaysInWidth guards the shift in ReverseMask: a reversal must
// not leak bits in from above the word.
func TestReverseMaskStaysInWidth(t *testing.T) {
	t.Parallel()

	if got, want := lfsr.ReverseMask[uint8](0x01), uint8(0x80); got != want {
		t.Errorf("ReverseMask[uint8](0x01) = %#x, want %#x", got, want)
	}

	if got, want := lfsr.ReverseMask[uint16](0x0001), uint16(0x8000); got != want {
		t.Errorf("ReverseMask[uint16](0x0001) = %#x, want %#x", got, want)
	}

	if got, want := lfsr.ReverseMask[uint32](0x00000001), uint32(0x80000000); got != want {
		t.Errorf("ReverseMask[uint32](0x00000001) = %#x, want %#x", got, want)
	}
}

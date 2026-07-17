package flfsr_test

import (
	"errors"
	"testing"

	"github.com/1995parham/LFSR.go/flfsr"
	"github.com/1995parham/LFSR.go/lfsr"
)

// Maximal-length tap masks, i.e. primitive feedback polynomials, for the widths
// exercised here.
const (
	// poly8 taps register bits 7, 5, 4, 3: x^8 + x^6 + x^5 + x^4 + 1.
	poly8 uint8 = 0xB8
	// poly16 taps register bits 15, 13, 12, 10: x^16 + x^14 + x^13 + x^11 + 1.
	poly16 uint16 = 0xB400
)

func TestNewRejectsDegenerateArguments(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		poly, seed uint8
		want       error
	}{
		"zero polynomial": {poly: 0x00, seed: 0x40, want: flfsr.ErrZeroPoly},
		"zero seed":       {poly: poly8, seed: 0x00, want: flfsr.ErrZeroSeed},
		"both zero":       {poly: 0x00, seed: 0x00, want: flfsr.ErrZeroPoly},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, err := flfsr.New(tc.poly, tc.seed)
			if !errors.Is(err, tc.want) {
				t.Fatalf("New(%#x, %#x) error = %v, want %v", tc.poly, tc.seed, err, tc.want)
			}

			if f != nil {
				t.Errorf("New(%#x, %#x) returned a register alongside an error", tc.poly, tc.seed)
			}
		})
	}
}

// TestNextMatchesReference pins the generic implementation against a hand
// written 8-bit register with the taps of poly8 hard coded, so a mistake in the
// tap-mask or parity handling cannot pass unnoticed.
func TestNextMatchesReference(t *testing.T) {
	t.Parallel()

	f, err := flfsr.New(poly8, uint8(0x40))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ref := uint8(0x40)

	for i := range 512 {
		out := f.Next()

		wantOut := ref >> 7
		ref = ref<<1 | (ref>>7^ref>>5^ref>>4^ref>>3)&0x01

		if out != wantOut {
			t.Fatalf("step %d: Next() = %d, want %d", i, out, wantOut)
		}

		if got := f.State(); got != ref {
			t.Fatalf("step %d: State() = %#08b, want %#08b", i, got, ref)
		}
	}
}

// TestMaximalPeriod is the property that matters for a pseudo-random source: a
// primitive polynomial must walk every non-zero state exactly once before it
// repeats.
func TestMaximalPeriod(t *testing.T) {
	t.Parallel()

	t.Run("8-bit", func(t *testing.T) {
		t.Parallel()

		f, err := flfsr.New(poly8, uint8(0x40))
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		if got, want := period(f.Next, f.State, 0x40), 1<<8-1; got != want {
			t.Errorf("period = %d, want %d", got, want)
		}
	})

	t.Run("16-bit", func(t *testing.T) {
		t.Parallel()

		f, err := flfsr.New(poly16, uint16(0xACE1))
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		if got, want := period(f.Next, f.State, 0xACE1), 1<<16-1; got != want {
			t.Errorf("period = %d, want %d", got, want)
		}
	})
}

// TestNoZeroState guards the invariant that keeps the register alive: a
// non-zero seed must never reach the all-zero fixed point.
func TestNoZeroState(t *testing.T) {
	t.Parallel()

	f, err := flfsr.New(poly8, uint8(0x01))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for i := range 1 << 10 {
		f.Next()

		if f.State() == 0 {
			t.Fatalf("register reached the all-zero state after %d steps", i+1)
		}
	}
}

// TestUintConsumesWholeWord checks that Uint is exactly its bit stream, packed
// most significant bit first, rather than a re-exposed register state.
func TestUintConsumesWholeWord(t *testing.T) {
	t.Parallel()

	packed, err := flfsr.New(poly8, uint8(0x40))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stream, err := flfsr.New(poly8, uint8(0x40))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for i := range 32 {
		var want uint8
		for range 8 {
			want = want<<1 | stream.Next()
		}

		if got := packed.Uint(); got != want {
			t.Fatalf("word %d: Uint() = %#08b, want %#08b", i, got, want)
		}
	}
}

// TestUintIsBalanced is a coarse smoke test for randomness: over a full period a
// maximal-length register emits 2^(n-1) ones and one fewer zero.
func TestUintIsBalanced(t *testing.T) {
	t.Parallel()

	f, err := flfsr.New(poly16, uint16(0xACE1))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var ones int
	for range 1<<16 - 1 {
		ones += int(f.Next())
	}

	if want := 1 << 15; ones != want {
		t.Errorf("ones over a full period = %d, want %d", ones, want)
	}
}

// TestUintClocksWidthBits instantiates every supported width and checks that
// Uint draws exactly as many bits as the word is wide. Getting this wrong on
// the wider types would silently shorten the stream, so assert it per width
// rather than trusting the 8-bit case to generalise.
func TestUintClocksWidthBits(t *testing.T) {
	t.Parallel()

	t.Run("8-bit", func(t *testing.T) {
		t.Parallel()
		assertUintClocks(t, poly8, uint8(0x40), 8)
	})

	t.Run("16-bit", func(t *testing.T) {
		t.Parallel()
		assertUintClocks(t, poly16, uint16(0xACE1), 16)
	})

	t.Run("32-bit", func(t *testing.T) {
		t.Parallel()
		assertUintClocks(t, uint32(0xA3000000), uint32(0xACE1ACE1), 32)
	})

	t.Run("64-bit", func(t *testing.T) {
		t.Parallel()
		assertUintClocks(t, uint64(0xD800000000000000), uint64(0xACE1ACE1ACE1ACE1), 64)
	})
}

// assertUintClocks checks that one Uint call leaves the register in the same
// state as exactly want single-bit clocks of an identically seeded register.
func assertUintClocks[W lfsr.Word](t *testing.T, poly, seed W, want int) {
	t.Helper()

	packed, err := flfsr.New(poly, seed)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stream, err := flfsr.New(poly, seed)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	packed.Uint()

	for range want {
		stream.Next()
	}

	if packed.State() != stream.State() {
		t.Errorf("Uint() did not clock exactly %d bits: state %#x, want %#x",
			want, packed.State(), stream.State())
	}
}

// period clocks the register until it returns to seed, reporting the number of
// steps taken, or limit+1 if it never does.
func period[W comparable](next func() uint8, state func() W, seed W) int {
	const limit = 1 << 17

	for i := 1; i <= limit; i++ {
		next()

		if state() == seed {
			return i
		}
	}

	return limit + 1
}

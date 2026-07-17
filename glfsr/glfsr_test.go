package glfsr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/1995parham/LFSR.go/flfsr"
	"github.com/1995parham/LFSR.go/glfsr"
	"github.com/1995parham/LFSR.go/lfsr"
)

// Galois tap masks, holding the low n coefficients of primitive polynomials.
// Each is the bit reversal of the flfsr mask for the same polynomial.
const (
	// poly8 is x^8 + x^4 + x^3 + x^2 + 1, which flfsr spells 0xB8.
	poly8 uint8 = 0x1D
	// poly16 is x^16 + x^5 + x^3 + x^2 + 1, which flfsr spells 0xB400.
	poly16 uint16 = 0x2D
)

func TestNewRejectsDegenerateArguments(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		poly, seed uint8
		want       error
	}{
		"zero polynomial": {poly: 0x00, seed: 0x40, want: glfsr.ErrZeroPoly},
		"zero seed":       {poly: poly8, seed: 0x00, want: glfsr.ErrZeroSeed},
		"both zero":       {poly: 0x00, seed: 0x00, want: glfsr.ErrZeroPoly},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			g, err := glfsr.New(tc.poly, tc.seed)
			if !errors.Is(err, tc.want) {
				t.Fatalf("New(%#x, %#x) error = %v, want %v", tc.poly, tc.seed, err, tc.want)
			}

			if g != nil {
				t.Errorf("New(%#x, %#x) returned a register alongside an error", tc.poly, tc.seed)
			}
		})
	}
}

// TestNextMatchesReference pins the implementation against a hand written 8-bit
// Galois register with the mask hard coded.
func TestNextMatchesReference(t *testing.T) {
	t.Parallel()

	g, err := glfsr.New(poly8, uint8(0x40))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ref := uint8(0x40)

	for i := range 512 {
		out := g.Next()

		wantOut := ref >> 7
		ref <<= 1

		if wantOut == 1 {
			ref ^= poly8
		}

		if out != wantOut {
			t.Fatalf("step %d: Next() = %d, want %d", i, out, wantOut)
		}

		if got := g.State(); got != ref {
			t.Fatalf("step %d: State() = %#08b, want %#08b", i, got, ref)
		}
	}
}

func TestMaximalPeriod(t *testing.T) {
	t.Parallel()

	t.Run("8-bit", func(t *testing.T) {
		t.Parallel()

		g, err := glfsr.New(poly8, uint8(0x40))
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		if got, want := period(g.Next, g.State, uint8(0x40)), 1<<8-1; got != want {
			t.Errorf("period = %d, want %d", got, want)
		}
	})

	t.Run("16-bit", func(t *testing.T) {
		t.Parallel()

		g, err := glfsr.New(poly16, uint16(0xACE1))
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		if got, want := period(g.Next, g.State, uint16(0xACE1)), 1<<16-1; got != want {
			t.Errorf("period = %d, want %d", got, want)
		}
	})
}

func TestNoZeroState(t *testing.T) {
	t.Parallel()

	g, err := glfsr.New(poly8, uint8(0x01))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for i := range 1 << 10 {
		g.Next()

		if g.State() == 0 {
			t.Fatalf("register reached the all-zero state after %d steps", i+1)
		}
	}
}

// TestAgreesWithFibonacci is the property that makes the two packages two views
// of one thing: given masks that denote the same polynomial, the Galois form
// emits the very sequence the Fibonacci form does, only entering it at a
// different phase. Checking for a rotation rather than a fixed offset keeps the
// test honest, since the phase depends on the seed.
func TestAgreesWithFibonacci(t *testing.T) {
	t.Parallel()

	const (
		fibMask uint8 = 0xB8
		seed    uint8 = 0x40
		cycle         = 1<<8 - 1
	)

	if got := lfsr.ReverseMask(fibMask); got != poly8 {
		t.Fatalf("ReverseMask(%#x) = %#x, want %#x", fibMask, got, poly8)
	}

	f, err := flfsr.New(fibMask, seed)
	if err != nil {
		t.Fatalf("flfsr.New: %v", err)
	}

	g, err := glfsr.New(lfsr.ReverseMask(fibMask), seed)
	if err != nil {
		t.Fatalf("glfsr.New: %v", err)
	}

	fib := stream(f.Next, cycle)
	gal := stream(g.Next, cycle)

	if fib == gal {
		t.Fatalf("the two forms are in phase, so this proves nothing about rotation")
	}

	// A rotation of fib is a substring of fib+fib.
	if !strings.Contains(fib+fib, gal) {
		t.Errorf("the galois stream is not a rotation of the fibonacci one\n fib = %s\n gal = %s", fib, gal)
	}
}

// TestWrongConventionIsCaught documents the trap the two conventions set: a mask
// meant for the other form is accepted, runs, and quietly gives a short period.
func TestWrongConventionIsCaught(t *testing.T) {
	t.Parallel()

	const fibMask uint8 = 0xB8 // the flfsr spelling; wrong here

	g, err := glfsr.New(fibMask, uint8(0x40))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := period(g.Next, g.State, uint8(0x40)); got == 1<<8-1 {
		t.Errorf("period = %d; the flfsr mask was expected to be non-primitive here, "+
			"so if it now reaches full period the conventions have converged and the "+
			"documentation needs revisiting", got)
	}
}

func stream(next func() uint8, count int) string {
	var b strings.Builder
	for range count {
		b.WriteByte('0' + next())
	}

	return b.String()
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

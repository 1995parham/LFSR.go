package glfsr_test

import (
	"testing"

	"github.com/1995parham/LFSR.go/glfsr"
	"github.com/1995parham/LFSR.go/lfsr"
)

// sink keeps the compiler from discarding the work under benchmark.
var (
	sinkBit  uint8
	sinkWord uint64
)

func BenchmarkNext(b *testing.B) {
	b.Run("8-bit", func(b *testing.B) {
		g, err := glfsr.New(poly8, uint8(0x40))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= g.Next()
		}

		sinkBit = s
	})

	b.Run("16-bit", func(b *testing.B) {
		g, err := glfsr.New(poly16, uint16(0xACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= g.Next()
		}

		sinkBit = s
	})

	b.Run("32-bit", func(b *testing.B) {
		g, err := glfsr.New(poly32, uint32(0xACE1ACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= g.Next()
		}

		sinkBit = s
	})

	b.Run("64-bit", func(b *testing.B) {
		g, err := glfsr.New(poly64, uint64(0xACE1ACE1ACE1ACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= g.Next()
		}

		sinkBit = s
	})
}

func BenchmarkUint(b *testing.B) {
	b.Run("8-bit", func(b *testing.B) {
		g, err := glfsr.New(poly8, uint8(0x40))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= g.Uint()
		}

		sinkWord = uint64(s)
	})

	b.Run("64-bit", func(b *testing.B) {
		g, err := glfsr.New(poly64, uint64(0xACE1ACE1ACE1ACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint64
		for b.Loop() {
			s ^= g.Uint()
		}

		sinkWord = s
	})
}

// branchy is the obvious spelling of a Galois clock, kept only so the benchmark
// can show why [glfsr.GLFSR.Next] does not use it. The outgoing bit is a coin
// flip, so this branch mispredicts about half the time.
type branchy struct{ data, poly uint32 }

func (b *branchy) Next() uint8 {
	out := uint8(b.data>>(lfsr.Width[uint32]()-1)) & 1

	b.data <<= 1
	if out == 1 {
		b.data ^= b.poly
	}

	return out
}

// BenchmarkNextBranchy measures the cost of the branch this package avoids. Run
// it against BenchmarkNext/32-bit to see the difference.
func BenchmarkNextBranchy(b *testing.B) {
	r := &branchy{data: 0xACE1ACE1, poly: poly32}

	var s uint8
	for b.Loop() {
		s ^= r.Next()
	}

	sinkBit = s
}

package flfsr_test

import (
	"testing"

	"github.com/1995parham/LFSR.go/flfsr"
)

// sink keeps the compiler from discarding the work under benchmark.
var (
	sinkBit  uint8
	sinkWord uint64
)

func BenchmarkNext(b *testing.B) {
	b.Run("8-bit", func(b *testing.B) {
		f, err := flfsr.New(poly8, uint8(0x40))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= f.Next()
		}

		sinkBit = s
	})

	b.Run("16-bit", func(b *testing.B) {
		f, err := flfsr.New(poly16, uint16(0xACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= f.Next()
		}

		sinkBit = s
	})

	b.Run("32-bit", func(b *testing.B) {
		f, err := flfsr.New(poly32, uint32(0xACE1ACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= f.Next()
		}

		sinkBit = s
	})

	b.Run("64-bit", func(b *testing.B) {
		f, err := flfsr.New(poly64, uint64(0xACE1ACE1ACE1ACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= f.Next()
		}

		sinkBit = s
	})
}

func BenchmarkUint(b *testing.B) {
	b.Run("8-bit", func(b *testing.B) {
		f, err := flfsr.New(poly8, uint8(0x40))
		if err != nil {
			b.Fatal(err)
		}

		var s uint8
		for b.Loop() {
			s ^= f.Uint()
		}

		sinkWord = uint64(s)
	})

	b.Run("64-bit", func(b *testing.B) {
		f, err := flfsr.New(poly64, uint64(0xACE1ACE1ACE1ACE1))
		if err != nil {
			b.Fatal(err)
		}

		var s uint64
		for b.Loop() {
			s ^= f.Uint()
		}

		sinkWord = s
	})
}

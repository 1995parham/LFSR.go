package prng_test

import (
	"fmt"
	"log"

	"github.com/1995parham/LFSR.go/flfsr"
	"github.com/1995parham/LFSR.go/glfsr"
	"github.com/1995parham/LFSR.go/prng"
)

// The shrinking generator the upstream C sample demonstrates: one register
// supplies the data, the other decides which of those bits survive.
func ExampleShrinking() {
	data, err := flfsr.New[uint8](0xB8, 0x40)
	if err != nil {
		log.Fatal(err)
	}

	clock, err := flfsr.New[uint8](0xB8, 0x17)
	if err != nil {
		log.Fatal(err)
	}

	g, err := prng.NewShrinking(data, clock)
	if err != nil {
		log.Fatal(err)
	}

	for range 16 {
		fmt.Print(g.Next())
	}

	fmt.Println()
	// Output: 0000001100001000
}

// The two registers need not share a form or a width.
func ExampleNewShrinking_mixed() {
	data, err := flfsr.New[uint8](0xB8, 0x40)
	if err != nil {
		log.Fatal(err)
	}

	clock, err := glfsr.New[uint16](0x2D, 0xACE1)
	if err != nil {
		log.Fatal(err)
	}

	g, err := prng.NewShrinking[uint8, uint16](data, clock)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#x %#x %#x\n", g.Uint(), g.Uint(), g.Uint())
	// Output: 0x4 0xc3 0x80
}

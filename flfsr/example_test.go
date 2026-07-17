package flfsr_test

import (
	"fmt"
	"log"

	"github.com/1995parham/LFSR.go/flfsr"
)

// Build an 8-bit register from the polynomial x^8 + x^4 + x^3 + x^2 + 1, whose
// taps on register bits 7, 5, 4 and 3 give the mask 0xB8, then read its bit
// stream.
func ExampleNew() {
	f, err := flfsr.New[uint8](0xB8, 0x40)
	if err != nil {
		log.Fatal(err)
	}

	for range 8 {
		fmt.Print(f.Next())
	}

	fmt.Println()
	// Output: 01000000
}

// The width of the type argument picks the register width, so the same code
// drives a 16-bit register built from x^16 + x^5 + x^3 + x^2 + 1.
//
// Note the first word: a register shifts its seed out before it shifts anything
// else, so the opening Uint of a freshly built register always reproduces the
// seed. Discard it if the seed is not meant to be observable.
func ExampleFLFSR_Uint() {
	f, err := flfsr.New[uint16](0xB400, 0xACE1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#04x %#04x %#04x\n", f.Uint(), f.Uint(), f.Uint())
	// Output: 0xace1 0xe455 0xdd17
}

package glfsr_test

import (
	"fmt"
	"log"

	"github.com/1995parham/LFSR.go/glfsr"
	"github.com/1995parham/LFSR.go/lfsr"
)

// Build an 8-bit register from the polynomial x^8 + x^4 + x^3 + x^2 + 1, whose
// low coefficients give the mask 0x1D, then read its bit stream.
func ExampleNew() {
	g, err := glfsr.New[uint8](0x1D, 0x40)
	if err != nil {
		log.Fatal(err)
	}

	for range 8 {
		fmt.Print(g.Next())
	}

	fmt.Println()
	// Output: 01000111
}

// A Galois mask is the bit reversal of the Fibonacci mask for the same
// polynomial, so a mask written for the flfsr package has to be converted
// before it means the same thing here.
func ExampleNew_fromFibonacciMask() {
	const fibonacci uint8 = 0xB8 // x^8 + x^4 + x^3 + x^2 + 1, as flfsr spells it

	g, err := glfsr.New(lfsr.ReverseMask(fibonacci), uint8(0x40))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#x\n", lfsr.ReverseMask(fibonacci))
	fmt.Printf("%#x %#x\n", g.Uint(), g.Uint())
	// Output:
	// 0x1d
	// 0x47 0x12
}

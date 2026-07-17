# LFSR.go

[![CI](https://github.com/1995parham/LFSR.go/actions/workflows/ci.yml/badge.svg)](https://github.com/1995parham/LFSR.go/actions/workflows/ci.yml)
[![Go Report](https://goreportcard.com/badge/github.com/1995parham/LFSR.go?style=flat-square)](https://goreportcard.com/report/github.com/1995parham/LFSR.go)
[![Go Reference](https://pkg.go.dev/badge/github.com/1995parham/LFSR.go.svg)](https://pkg.go.dev/github.com/1995parham/LFSR.go)

## Introduction

Go implementation of the linear feedback shift register, generic over register
width, based on [Michael Foukarakis' LFSR implementation and tools](https://github.com/mfukar/lfsr).

## LFSR ?

In computing, a linear-feedback shift register (LFSR) is a shift register whose input bit
is a linear function of its previous state.
Applications of LFSRs include generating pseudo-random numbers, pseudo-noise sequences,
fast digital counters, and whitening sequences. Both hardware and software implementations of LFSRs are common.

## Usage

```go
import "github.com/1995parham/LFSR.go/flfsr"

// x^8 + x^6 + x^5 + x^4 + 1 taps register bits 7, 5, 4 and 3.
f, err := flfsr.New[uint8](0xB8, 0x40)
if err != nil {
    return err
}

f.Next()  // one pseudo-random bit, 0 or 1
f.Uint()  // eight clocks packed into a byte
f.State() // the raw register, without clocking it
```

The type argument picks the register width; `uint8`, `uint16`, `uint32` and
`uint64` are supported. `New` rejects a zero polynomial and a zero seed, both of
which collapse the register to the all-zero state.

Read the bit stream through `Next`, or whole words through `Uint`. Do not treat
successive `State` values as random: they are the same bits shifted along, so
consecutive states overlap in all but one bit.

### Tap masks

`poly` is a mask over **register bit positions**, not a table of polynomial
exponents. Setting bit `i` taps register bit `i`, which stands for the term
`x^(i+1)`. The polynomial's implicit `+1` term is the incoming feedback bit and
never appears in the mask.

| Polynomial                      | Taps (register bits) | Mask     |
| ------------------------------- | -------------------- | -------- |
| `x^8 + x^6 + x^5 + x^4 + 1`     | 7, 5, 4, 3           | `0xB8`   |
| `x^16 + x^14 + x^13 + x^11 + 1` | 15, 13, 12, 10       | `0xB400` |

Both masks above are primitive, so they reach the maximal period of `2^n - 1`.
`New` only rejects the degenerate masks; it does not verify primitivity, and a
non-primitive polynomial simply gives a shorter period.

### Fibonacci LFSR

`flfsr` implements the Fibonacci form: the register shifts towards its most
significant bit, emitting the outgoing top bit, and feeds the parity of the
tapped bits back into bit 0. The reference implementation for
`x^16 + x^14 + x^13 + x^11 + 1`:

```c
# include <stdint.h>
int main(void)
{
    uint16_t start_state = 0xACE1u;  /* Any nonzero start state will work. */
    uint16_t lfsr = start_state;
    uint16_t bit;                    /* Must be 16bit to allow bit<<15 later in the code */
    unsigned period = 0;

    do
    {
        /* taps: 16 14 13 11; feedback polynomial: x^16 + x^14 + x^13 + x^11 + 1 */
        bit  = ((lfsr >> 0) ^ (lfsr >> 2) ^ (lfsr >> 3) ^ (lfsr >> 5) ) & 1;
        lfsr =  (lfsr >> 1) | (bit << 15);
        ++period;
    } while (lfsr != start_state);

    return 0;
}
```

Note that this snippet shifts right and taps the low bits, the mirror image of
the convention used here.

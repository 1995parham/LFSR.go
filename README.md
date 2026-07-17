
<h1 align="center">
    LFSR.go
</h1>

<p align="center">
    <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/1995parham/LFSR.go/ci.yml?style=for-the-badge&logo=github">
</p>

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

// x^8 + x^4 + x^3 + x^2 + 1 taps register bits 7, 5, 4 and 3.
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
exponents. Because the register shifts up, mask bit `i` stands for the term
`x^(n-1-i)`, where `n` is the register width. The `x^n` term is implicit — it is
the feedback itself — and the polynomial's constant term lands on bit `n-1`.
Every primitive polynomial has a constant term, so **the top bit of a usable
mask is always set**.

These four are primitive, so each reaches the maximal period of `2^n - 1`:

| Width | Polynomial                    | Taps (register bits) | Mask                 |
| ----- | ----------------------------- | -------------------- | -------------------- |
| 8     | `x^8 + x^4 + x^3 + x^2 + 1`   | 7, 5, 4, 3           | `0xB8`               |
| 16    | `x^16 + x^5 + x^3 + x^2 + 1`  | 15, 13, 12, 10       | `0xB400`             |
| 32    | `x^32 + x^22 + x^2 + x + 1`   | 31, 30, 29, 9        | `0xE0000200`         |
| 64    | `x^64 + x^63 + x^61 + x^60 + 1` | 63, 3, 2, 0        | `0x800000000000000D` |

`New` only rejects the degenerate masks; it does not verify primitivity, and a
non-primitive polynomial simply gives a shorter period.

#### Borrowing masks from published tables

Most published tap tables — [Wikipedia's LFSR article][wiki], Xilinx
[XAPP052][], and Philip Koopman's [Maximal Length LFSR Feedback Terms][koopman],
which covers every width from 4 to 64 bits — describe **right-shifting**
registers, where mask bit `i` means `x^(i+1)`. A mask lifted from one of those
and used here yields the *reciprocal* of the polynomial the table names.

That is harmless in practice: the reciprocal of a primitive polynomial is itself
primitive and has the same period, so you still get `2^n - 1`. But the sequence
is not the one the table describes, and the polynomial in the table is not the
polynomial you are running. The 8- and 16-bit rows above are exactly this case —
`0xB8` and `0xB400` are the familiar masks for `x^8 + x^6 + x^5 + x^4 + 1` and
`x^16 + x^14 + x^13 + x^11 + 1`, and here they run those polynomials' reciprocals.

To run a polynomial a table names, reverse its low `n` coefficients: `x^32 + x^22
+ x^2 + x + 1` becomes mask `0xE0000200`, not the `0x80200003` a right-shifting
implementation would use.

[wiki]: https://en.wikipedia.org/wiki/Linear-feedback_shift_register
[XAPP052]: https://docs.amd.com/v/u/en-US/xapp052
[koopman]: https://users.ece.cmu.edu/~koopman/lfsr/index.html

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

// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program generates a .s file using a pseudorandom
// value stream for the runtime function data.
// The pclntab test checks that the linked copy
// still has the same pseudorandom value stream.

package main

import (
	"fmt"
	"math/rand"
)

func main() {
	fmt.Printf("// generated by genpcln.go; do not edit\n\n")
	for f := 0; f < 3; f++ {
		r := rand.New(rand.NewSource(int64(f)))
		file := "input"
		line := 1
		args := r.Intn(100) * 8
		frame := 32 + r.Intn(32)/8*8
		fmt.Printf("#line %d %q\n", line, file)
		fmt.Printf("TEXT func%d(SB),7,$%d-%d\n", f, frame, args)
		fmt.Printf("\tFUNCDATA $1, funcdata%d(SB)\n", f)
		fmt.Printf("#line %d %q\n", line, file)
		size := 200 + r.Intn(100)*8
		spadj := 0
		flushed := 0
		firstpc := 4
		flush := func(i int) {
			for i-flushed >= 10 {
				fmt.Printf("#line %d %q\n", line, file)
				fmt.Printf("/*%#04x*/\tMOVQ $0x123456789, AX\n", firstpc+flushed)
				flushed += 10
			}
			for i-flushed >= 5 {
				fmt.Printf("#line %d %q\n", line, file)
				fmt.Printf("/*%#04x*/\tMOVL $0x1234567, AX\n", firstpc+flushed)
				flushed += 5
			}
			for i-flushed > 0 {
				fmt.Printf("#line %d %q\n", line, file)
				fmt.Printf("/*%#04x*/\tBYTE $0\n", firstpc+flushed)
				flushed++
			}
		}
		for i := 0; i < size; i++ {
			// Possible SP adjustment.
			if r.Intn(100) == 0 {
				flush(i)
				fmt.Printf("#line %d %q\n", line, file)
				if spadj <= -32 || spadj < 32 && r.Intn(2) == 0 {
					spadj += 8
					fmt.Printf("/*%#04x*/\tPUSHQ AX\n", firstpc+i)
				} else {
					spadj -= 8
					fmt.Printf("/*%#04x*/\tPOPQ AX\n", firstpc+i)
				}
				i += 1
				flushed = i
			}

			// Possible PCFile change.
			if r.Intn(100) == 0 {
				flush(i)
				file = fmt.Sprintf("file%d.s", r.Intn(10))
				line = r.Intn(100) + 1
			}

			// Possible PCLine change.
			if r.Intn(10) == 0 {
				flush(i)
				line = r.Intn(1000) + 1
			}

			// Possible PCData $1 change.
			if r.Intn(100) == 0 {
				flush(i)
				fmt.Printf("/*%6s*/\tPCDATA $1, $%d\n", "", r.Intn(1000))
			}

			// Possible PCData $2 change.
			if r.Intn(100) == 0 {
				flush(i)
				fmt.Printf("/*%6s*/\tPCDATA $2, $%d\n", "", r.Intn(1000))
			}
		}
		flush(size)
		for spadj < 0 {
			fmt.Printf("\tPUSHQ AX\n")
			spadj += 8
		}
		for spadj > 0 {
			fmt.Printf("\tPOPQ AX\n")
			spadj -= 8
		}
		fmt.Printf("\tRET\n")

		fmt.Printf("\n")
		fmt.Printf("GLOBL funcdata%d(SB), $16\n", f)
	}

	fmt.Printf("\nTEXT start(SB),7,$0\n")
	for f := 0; f < 3; f++ {
		fmt.Printf("\tCALL func%d(SB)\n", f)
	}
	fmt.Printf("\tMOVQ $runtime·pclntab(SB), AX\n")
	fmt.Printf("\n\tRET\n")
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

// addr (add register) stores into register C the result of adding register A and register B.
func addr(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] + r[b]
	return r
}

// addi (add immediate) stores into register C the result of adding register A and value B.
func addi(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] + b
	return r
}

// mulr (multiply register) stores into register C the result of multiplying register A and register B.
func mulr(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] * r[b]
	return r
}

// muli (multiply immediate) stores into register C the result of multiplying register A and value B.
func muli(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] * b
	return r
}

// banr (bitwise AND register) stores into register C the result of the bitwise AND of register A and register B.
func banr(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] & r[b]
	return r
}

// bani (bitwise AND immediate) stores into register C the result of the bitwise AND of register A and value B.
func bani(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] & b
	return r
}

// borr (bitwise OR register) stores into register C the result of the bitwise OR of register A and register B.
func borr(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] | r[b]
	return r
}

// bori (bitwise OR immediate) stores into register C the result of the bitwise OR of register A and value B.
func bori(r [4]int, a, b, c int) [4]int {
	r[c] = r[a] | b
	return r
}

// setr (set register) copies the contents of register A into register C. (Input B is ignored.)
func setr(r [4]int, a, b, c int) [4]int {
	r[c] = r[a]
	return r
}

// seti (set immediate) stores value A into register C. (Input B is ignored.)
func seti(r [4]int, a, b, c int) [4]int {
	r[c] = a
	return r
}

// gtir (greater-than immediate/register) sets register C to 1 if value A is greater than register B. Otherwise, register C is set to 0.
func gtir(r [4]int, a, b, c int) [4]int {
	if a > r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// gtri (greater-than register/immediate) sets register C to 1 if register A is greater than value B. Otherwise, register C is set to 0.
func gtri(r [4]int, a, b, c int) [4]int {
	if r[a] > b {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// gtrr (greater-than register/register) sets register C to 1 if register A is greater than register B. Otherwise, register C is set to 0.
func gtrr(r [4]int, a, b, c int) [4]int {
	if r[a] > r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// eqir (equal immediate/register) sets register C to 1 if value A is equal to register B. Otherwise, register C is set to 0.
func eqir(r [4]int, a, b, c int) [4]int {
	if a == r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// eqri (equal register/immediate) sets register C to 1 if register A is equal to value B. Otherwise, register C is set to 0.
func eqri(r [4]int, a, b, c int) [4]int {
	if r[a] == b {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// eqrr (equal register/register) sets register C to 1 if register A is equal to register B. Otherwise, register C is set to 0.
func eqrr(r [4]int, a, b, c int) [4]int {
	if r[a] == r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

type opFunc func(r [4]int, a, b, c int) [4]int

var (
	ops = [16]opFunc{
		addr, addi,
		mulr, muli,
		banr, bani,
		borr, bori,
		setr, seti,
		gtir, gtri, gtrr,
		eqir, eqri, eqrr,
	}
	names = [16]string{
		"addr", "addi",
		"mulr", "muli",
		"banr", "bani",
		"borr", "bori",
		"setr", "seti",
		"gtir", "gtri", "gtrr",
		"eqir", "eqri", "eqrr",
	}
)

type sample struct {
	before [4]int
	op     [4]int
	after  [4]int
}

func run(in, out *os.File) error {
	samples, err := read(in)
	if err != nil {
		return err
	}

	// part 1
	log.Printf("read %v samples", len(samples))
	threeOrMore := 0
	for _, sample := range samples {
		// log.Printf("sample[%v]: %v", i, sample)
		n := 0
		for _, op := range ops {
			after := op(sample.before, sample.op[1], sample.op[2], sample.op[3])
			eq := true
			for i, rv := range after {
				eq = eq && rv == sample.after[i]
			}
			if eq {
				// log.Printf("could've been %s", names[j])
				n++
			}
		}
		// log.Printf("could've been %v ops", n)
		if n >= 3 {
			threeOrMore++
		}
	}
	log.Printf("%v samples could've been 3 or more ops", threeOrMore)

	// part 2
	// TODO

	return nil
}

var (
	beforePat = regexp.MustCompile(`Before: +\[(\d+), (\d+), (\d+), (\d+)\]`)
	opPat     = regexp.MustCompile(`(\d+) (\d+) (\d+) (\d+)`)
	afterPat  = regexp.MustCompile(`After: +\[(\d+), (\d+), (\d+), (\d+)\]`)
)

func read(r io.Reader) (samples []sample, err error) {
	sc := bufio.NewScanner(r)
	samples, err = scanSamples(sc)
	// if err == nil {
	// 	TODO parse trailer program
	// }
	return samples, err
}

func scanSamples(sc *bufio.Scanner) (samples []sample, _ error) {
	lineNum := 0
	for done := false; sc.Err() == nil && !done; {
		if err := func() error {
			var s sample
			n := 0
			for _, step := range []func() error{
				// Before: [r1, r2, r3, r4]
				func() error {
					if sc.Text() == "" {
						// sampples end with a blank line where a before would be
						done = true
						return nil
					}
					return expect(sc.Text(), beforePat, func(parts []string) error {
						n++
						return parseInts(s.before[:], parts[1:])
					})
				},
				// OP A B C
				func() error {
					return expect(sc.Text(), opPat, func(parts []string) error {
						n++
						return parseInts(s.op[:], parts[1:])
					})
				},
				// After:  [r1, r2, r3, r4]
				func() error {
					return expect(sc.Text(), afterPat, func(parts []string) error {
						n++
						err := parseInts(s.after[:], parts[1:])
						if err == nil {
							samples = append(samples, s)
						}
						return err
					})
				},
				func() error {
					if sc.Text() == "" {
						return nil
					}
					return fmt.Errorf("expecting blank")
				},
			} {
				if done {
					break
				}
				if !sc.Scan() {
					done = true
					break
				}
				lineNum++
				if err := step(); err != nil {
					return err
				}
			}
			if n > 0 && n < 3 {
				return fmt.Errorf("expected all 3 'before', 'op', and 'after' lines, saw only %v", n)
			}
			return nil
		}(); err != nil {
			return samples, fmt.Errorf("%v (in line %d:%q)", err, lineNum, sc.Text())
		}
	}

	return samples, sc.Err()
}

func expect(s string, pat *regexp.Regexp, f func([]string) error) error {
	parts := pat.FindStringSubmatch(s)
	if len(parts) == 0 {
		return fmt.Errorf("expecting %v", pat)
	}
	return f(parts)
}

func parseInts(ns []int, parts []string) error {
	i := 0
	for i < len(ns) && i < len(parts) {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return fmt.Errorf("failed to parse int %q: %v", parts[i], err)
		}
		ns[i] = n
		i++
	}
	if i < len(ns) || i < len(parts) {
		return fmt.Errorf("expected %v integers, got %v", len(ns), len(parts))
	}
	return nil
}

package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/jcorbin/anansi"
)

var verbose = flag.Bool("v", false, "verbose logs")

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

// addr (add register) stores into register C the result of adding register A and register B.
func addr(r regs, a, b, c int) regs {
	r[c] = r[a] + r[b]
	return r
}

// addi (add immediate) stores into register C the result of adding register A and value B.
func addi(r regs, a, b, c int) regs {
	r[c] = r[a] + b
	return r
}

// mulr (multiply register) stores into register C the result of multiplying register A and register B.
func mulr(r regs, a, b, c int) regs {
	r[c] = r[a] * r[b]
	return r
}

// muli (multiply immediate) stores into register C the result of multiplying register A and value B.
func muli(r regs, a, b, c int) regs {
	r[c] = r[a] * b
	return r
}

// banr (bitwise AND register) stores into register C the result of the bitwise AND of register A and register B.
func banr(r regs, a, b, c int) regs {
	r[c] = r[a] & r[b]
	return r
}

// bani (bitwise AND immediate) stores into register C the result of the bitwise AND of register A and value B.
func bani(r regs, a, b, c int) regs {
	r[c] = r[a] & b
	return r
}

// borr (bitwise OR register) stores into register C the result of the bitwise OR of register A and register B.
func borr(r regs, a, b, c int) regs {
	r[c] = r[a] | r[b]
	return r
}

// bori (bitwise OR immediate) stores into register C the result of the bitwise OR of register A and value B.
func bori(r regs, a, b, c int) regs {
	r[c] = r[a] | b
	return r
}

// setr (set register) copies the contents of register A into register C. (Input B is ignored.)
func setr(r regs, a, b, c int) regs {
	r[c] = r[a]
	return r
}

// seti (set immediate) stores value A into register C. (Input B is ignored.)
func seti(r regs, a, b, c int) regs {
	r[c] = a
	return r
}

// gtir (greater-than immediate/register) sets register C to 1 if value A is greater than register B. Otherwise, register C is set to 0.
func gtir(r regs, a, b, c int) regs {
	if a > r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// gtri (greater-than register/immediate) sets register C to 1 if register A is greater than value B. Otherwise, register C is set to 0.
func gtri(r regs, a, b, c int) regs {
	if r[a] > b {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// gtrr (greater-than register/register) sets register C to 1 if register A is greater than register B. Otherwise, register C is set to 0.
func gtrr(r regs, a, b, c int) regs {
	if r[a] > r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// eqir (equal immediate/register) sets register C to 1 if value A is equal to register B. Otherwise, register C is set to 0.
func eqir(r regs, a, b, c int) regs {
	if a == r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// eqri (equal register/immediate) sets register C to 1 if register A is equal to value B. Otherwise, register C is set to 0.
func eqri(r regs, a, b, c int) regs {
	if r[a] == b {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

// eqrr (equal register/register) sets register C to 1 if register A is equal to register B. Otherwise, register C is set to 0.
func eqrr(r regs, a, b, c int) regs {
	if r[a] == r[b] {
		r[c] = 1
	} else {
		r[c] = 0
	}
	return r
}

type opFunc func(r regs, a, b, c int) regs

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

type regs [4]int
type inst [4]int

func (r regs) eq(other regs) bool {
	for i, rv := range r {
		if rv != other[i] {
			return false
		}
	}
	return true
}

type sample struct {
	before regs
	op     inst
	after  regs
}

func run(in, out *os.File) error {
	prog, samples, err := read(in)
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
	var srch search
	codes, found, err := srch.run(samples)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("found no working opcode mapping")
	}

	var buf bytes.Buffer
	for code, op := range codes {
		fmt.Fprintf(&buf, " %v:%v", code, names[op])
	}
	log.Printf("FOUND%s", buf.Bytes())

	// run the program
	step := 0
	var rs regs
	log.Printf("%v: %v", step, rs)
	for _, in := range prog {
		step++
		code, a, b, c := in[0], in[1], in[2], in[3]
		op := codes[code]
		opFunc := ops[op]
		rs = opFunc(rs, a, b, c)
		log.Printf("%v: %s %v %v %v => %v", step, names[op], a, b, c, rs)
	}

	return nil
}

var (
	beforePat = regexp.MustCompile(`Before: +\[(\d+), (\d+), (\d+), (\d+)\]`)
	opPat     = regexp.MustCompile(`(\d+) (\d+) (\d+) (\d+)`)
	afterPat  = regexp.MustCompile(`After: +\[(\d+), (\d+), (\d+), (\d+)\]`)
)

func read(r io.Reader) (prog []inst, samples []sample, err error) {
	sc := bufio.NewScanner(r)
	samples, err = scanSamples(sc)
	if err == nil {
		prog, err = scanProgram(sc)
	}
	return prog, samples, err
}

func scanSamples(sc *bufio.Scanner) (samples []sample, _ error) {
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
				if err := step(); err != nil {
					return err
				}
			}
			if n > 0 && n < 3 {
				return fmt.Errorf("expected all 3 'before', 'op', and 'after' lines, saw only %v", n)
			}
			return nil
		}(); err != nil {
			return samples, fmt.Errorf("%v (in line %q)", err, sc.Text())
		}
	}

	return samples, sc.Err()
}

func scanProgram(sc *bufio.Scanner) (prog []inst, _ error) {
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			if len(prog) == 0 {
				continue
			}
			return prog, nil
		}
		var next inst
		if err := expect(line, opPat, func(parts []string) error {
			return parseInts(next[:], parts[1:])
		}); err != nil {
			return prog, fmt.Errorf("%v (in line %q)", err, sc.Text())
		}
		prog = append(prog, next)
	}
	return prog, sc.Err()
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

type search struct {
	st    []searchState
	depth []int
}

type searchEl struct {
	depth int
	searchState
}

func (srch *search) run(samples []sample) (codes [16]int, found bool, err error) {
	srch.init(1024, searchState{samples: samples})
	for i := 0; srch.expand(func(res searchState) {
		if found {
			err = errors.New("non-unique result")
		} else {
			found, codes = true, res.codes
		}
	}); i++ {
		if i >= 10000 {
			return codes, found, errors.New("search limit exceeded")
		}
	}
	return codes, found, err
}

func (srch *search) init(n int, st ...searchState) {
	srch.st = make([]searchState, 0, n)
	srch.st = append(srch.st, st...)
	srch.depth = make([]int, len(srch.st), n)
}

func (srch *search) expand(emit func(searchState)) bool {
	if len(srch.depth) == 0 {
		return false
	}

	depth, st := srch.pop()
	any := false
	st.expand(
		func(newST searchState) {
			srch.st = append(srch.st, newST)
			srch.depth = append(srch.depth, depth+1)
			any = true
		},
		emit,
	)
	if any {
		heap.Init(srch)
	}
	return true
}

func (srch *search) Len() int               { return len(srch.depth) }
func (srch *search) Less(i int, j int) bool { return srch.depth[i] > srch.depth[j] }
func (srch *search) Swap(i int, j int) {
	srch.st[i], srch.st[j] = srch.st[j], srch.st[i]
	srch.depth[i], srch.depth[j] = srch.depth[j], srch.depth[i]
}

func (srch *search) Push(x interface{}) {
	el := x.(searchEl)
	srch.st = append(srch.st, el.searchState)
	srch.depth = append(srch.depth, el.depth)
}

func (srch *search) Pop() interface{} {
	i := len(srch.st) - 1
	if i < 0 {
		return nil
	}
	st := srch.st[i]
	depth := srch.depth[i]
	srch.st = srch.st[:i]
	srch.depth = srch.depth[:i]
	return searchEl{depth, st}
}

func (srch *search) pop() (int, searchState) {
	x := heap.Pop(srch)
	el, ok := x.(searchEl)
	if !ok {
		return -1, searchState{}
	}
	return el.depth, el.searchState
}

type searchState struct {
	samples []sample

	knownCode [16]bool
	codes     [16]int

	knownOp [16]bool
	ops     [16]int
}

type searchStateScratch struct {
	searchState
	counts   opCodeSample
	obviOp   [16]int
	obviCode [16]int
}

func (st *searchState) countSamples() (obs opCodeSample) {
	for _, sample := range st.samples {
		if code := sample.op[0]; !st.knownOp[code] {
			a, b, c := sample.op[1], sample.op[2], sample.op[3]
			for op, opFunc := range ops {
				if !st.knownCode[op] {
					if after := opFunc(sample.before, a, b, c); after.eq(sample.after) {
						obs[op][code]++
					}
				}
			}
		}
	}
	return obs
}

func (st *searchState) assign(op, code int) {
	st.knownCode[op] = true
	st.codes[op] = code

	st.knownOp[code] = true
	st.ops[code] = op
}

func (st searchState) expand(push func(searchState), emit func(searchState)) {
	tmp := searchStateScratch{searchState: st}
	tmp.expand(push, emit)
}

func (st *searchStateScratch) expand(push func(searchState), emit func(searchState)) {
	st.counts = st.countSamples()
	st.collectObviousCounts()

	var buf bytes.Buffer

	nObviOps, nObviCodes, ok := st.checkObvi(&buf)
	if !ok {
		st.logf("FAIL%s", buf.Bytes())
		return
	}

	if nObviOps > 0 || nObviCodes > 0 {
		if nObviOps > 0 {
			for op, code := range st.obviOp {
				if code >= 0 {
					st.assign(op, code)
					fmt.Fprintf(&buf, " %v:%v", code, names[op])
				}
			}
		}
		if nObviCodes > 0 {
			for code, op := range st.obviCode {
				if op >= 0 && !st.knownCode[op] {
					st.assign(op, code)
					fmt.Fprintf(&buf, " %v:%v", code, names[op])
				}
			}
		}

		st.counts = st.countSamples()
		if *verbose {
			st.logf("assigned%s", buf.Bytes())
		}

		all := true
		for _, known := range st.knownCode {
			all = all && known
		}
		if all {
			emit(st.searchState)
		} else {
			push(st.searchState)
		}

		return
	}

	for op, code := range st.obviOp {
		if code >= 0 {
			fmt.Fprintf(&buf, " obviOp[%v]:%v", names[op], code)
		}
	}
	for code, op := range st.obviCode {
		if op >= 0 {
			fmt.Fprintf(&buf, " obviCode[%v]:%v", code, names[op])
		}
	}
	st.logf(":shrug:%s", buf.Bytes())
}

func (st *searchStateScratch) collectObviousCounts() {
	for i := range st.obviOp {
		st.obviOp[i] = -1
	}
	for i := range st.obviCode {
		st.obviCode[i] = -1
	}
	for op, counts := range st.counts {
		only := -1
		for code, n := range counts {
			if n > 0 {
				switch st.obviCode[code] {
				case -1:
					st.obviCode[code] = op
				default:
					st.obviCode[code] = -2
				}
				switch only {
				case -1:
					only = code
				default:
					only = -2
				}
			}
		}
		if only >= 0 {
			st.obviOp[op] = only
		}
	}
}

func (st *searchStateScratch) checkObvi(buf *bytes.Buffer) (nObviOps, nObviCodes int, ok bool) {
	ok = true

	for op, code := range st.obviOp {
		if code < 0 {
			continue
		}
		nObviOps++
		if st.obviCode[code] >= 0 {
			if st.knownOp[code] && st.ops[code] != op {
				if buf == nil {
					return 0, 0, false
				}
				fmt.Fprintf(buf, " conflict(obviOp[%v] != ops[%v])", code, st.ops[code])
			} else if st.obviCode[code] != op {
				if buf != nil {
					return 0, 0, false
				}
				fmt.Fprintf(buf, " conflict(obviOp[%v] != obviCode[%v])", code, st.obviCode[code])
			}
		}
	}

	for code, op := range st.obviCode {
		if op < 0 {
			continue
		}
		nObviCodes++
		if st.obviOp[op] >= 0 {
			if st.knownOp[code] && st.ops[code] != op {
				if buf == nil {
					return 0, 0, false
				}
				fmt.Fprintf(buf, " conflict(obviCode[%v] != ops[%v])", code, st.ops[code])
			} else if st.obviOp[op] != code {
				if buf != nil {
					return 0, 0, false
				}
				fmt.Fprintf(buf, " conflict(obviCode[%v]) != obviOp[%v]", op, st.obviOp[op])
			}
		}
	}

	return nObviOps, nObviCodes, ok
}

func (st *searchStateScratch) logf(mess string, args ...interface{}) {
	if len(mess) > 0 {
		log.Printf(mess, args...)
	}

	var buf bytes.Buffer

	buf.Reset()
	buf.WriteString("known:[")
	first := true
	for code, known := range st.knownOp {
		if known {
			if first {
				first = false
			} else {
				buf.WriteByte(' ')
			}
			fmt.Fprintf(&buf, "%d=%s", code, names[st.ops[code]])
		}
	}
	buf.WriteString("]")
	log.Printf(buf.String())

	buf.Reset()
	buf.WriteString("unknown:[")
	first = true
	for code, known := range st.knownOp {
		if !known {
			if first {
				first = false
			} else {
				buf.WriteByte(' ')
			}
			fmt.Fprintf(&buf, "%d", code)
		}
	}
	buf.WriteString("]")
	log.Printf(buf.String())

	st.counts.log()
}

type opCodeSample [16][16]int

func (obs opCodeSample) log() {
	var lines [17]bytes.Buffer
	lines[0].WriteString("     ")
	for code := 0; code < 16; code++ {
		fmt.Fprintf(&lines[code+1], "% 3d: ", code)
	}
	for i, counts := range obs {
		lines[0].WriteByte(' ')
		w, _ := lines[0].WriteString(names[i])
		for j, n := range counts {
			lines[1+j].WriteByte(' ')
			fmt.Fprintf(&lines[1+j], "% *d", w, n)
		}
	}
	for i := range lines {
		log.Printf(lines[i].String())
	}
}

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jcorbin/anansi"
)

var tap = taps{}

var (
	verbose = flag.Bool("v", false, "verbose logs")
	initR0  = flag.Int("r0", 0, "initial register 0")
	initR1  = flag.Int("r1", 0, "initial register 1")
	initR2  = flag.Int("r2", 0, "initial register 2")
	initR3  = flag.Int("r3", 0, "initial register 3")
	initR4  = flag.Int("r4", 0, "initial register 4")
	initR5  = flag.Int("r5", 0, "initial register 5")
	limit   = flag.Int("limit", 0, "limit number of operations")
)

func main() {
	flag.Var(tap, "tap", "add a debug tap")
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func addr(r *regs, a, b, c int) { r[c] = r[a] + r[b] }
func addi(r *regs, a, b, c int) { r[c] = r[a] + b }
func mulr(r *regs, a, b, c int) { r[c] = r[a] * r[b] }
func muli(r *regs, a, b, c int) { r[c] = r[a] * b }
func banr(r *regs, a, b, c int) { r[c] = r[a] & r[b] }
func bani(r *regs, a, b, c int) { r[c] = r[a] & b }
func borr(r *regs, a, b, c int) { r[c] = r[a] | r[b] }
func bori(r *regs, a, b, c int) { r[c] = r[a] | b }
func setr(r *regs, a, b, c int) { r[c] = r[a] }
func seti(r *regs, a, b, c int) { r[c] = a }
func gtir(r *regs, a, b, c int) { r[c] = btoi(a > r[b]) }
func gtri(r *regs, a, b, c int) { r[c] = btoi(r[a] > b) }
func gtrr(r *regs, a, b, c int) { r[c] = btoi(r[a] > r[b]) }
func eqir(r *regs, a, b, c int) { r[c] = btoi(a == r[b]) }
func eqri(r *regs, a, b, c int) { r[c] = btoi(r[a] == b) }
func eqrr(r *regs, a, b, c int) { r[c] = btoi(r[a] == r[b]) }

type opFunc func(r *regs, a, b, c int)

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
	name2opCode = map[string]int{
		"addr": 0,
		"addi": 1,
		"mulr": 2,
		"muli": 3,
		"banr": 4,
		"bani": 5,
		"borr": 6,
		"bori": 7,
		"setr": 8,
		"seti": 9,
		"gtir": 10,
		"gtri": 11,
		"gtrr": 12,
		"eqir": 13,
		"eqri": 14,
		"eqrr": 15,
	}
)

type regs [6]int
type inst [4]int
type comp struct {
	op      opFunc
	a, b, c int
}
type program []inst

func (r regs) eq(other regs) bool {
	for i, rv := range r {
		if rv != other[i] {
			return false
		}
	}
	return true
}

func (prog program) run(ipReg int, vm regs, lim int) (n int, ok bool) {
	ip := &vm[ipReg]

	if *verbose {
		// slow interpretor when logging verbose
		var buf bytes.Buffer
		for i := 0; *ip < len(prog); i++ {
			n++
			if lim > 0 && i >= lim {
				return n, false
			}
			in := prog[*ip]
			op, a, b, c := in[0], in[1], in[2], in[3]
			buf.Reset()
			fmt.Fprintf(&buf, "ip=%v %v %s %v %v %v", *ip, vm, names[op], a, b, c)
			ops[op](&vm, a, b, c)
			fmt.Fprintf(&buf, " %v", vm)
			log.Printf(buf.String())
			*ip++
		}
	} else {
		// faster interpretor when just barging through
		t0 := time.Now()
		i := 0
		for ; *ip < len(prog); i++ {
			n++
			if tapFn := tap[*ip]; tapFn != nil {
				tapFn(n, *ip, vm)
			}
			if lim > 0 && i >= lim {
				return n, false
			}
			in := prog[*ip]
			switch op, a, b, c := in[0], in[1], in[2], in[3]; op {
			case 0:
				vm[c] = vm[a] + vm[b]
			case 1:
				vm[c] = vm[a] + b
			case 2:
				vm[c] = vm[a] * vm[b]
			case 3:
				vm[c] = vm[a] * b
			case 4:
				vm[c] = vm[a] & vm[b]
			case 5:
				vm[c] = vm[a] & b
			case 6:
				vm[c] = vm[a] | vm[b]
			case 7:
				vm[c] = vm[a] | b
			case 8:
				vm[c] = vm[a]
			case 9:
				vm[c] = a
			case 10:
				vm[c] = btoi(a > vm[b])
			case 11:
				vm[c] = btoi(vm[a] > b)
			case 12:
				vm[c] = btoi(vm[a] > vm[b])
			case 13:
				vm[c] = btoi(a == vm[b])
			case 14:
				vm[c] = btoi(vm[a] == b)
			case 15:
				vm[c] = btoi(vm[a] == vm[b])
			}
			*ip++
		}
		t1 := time.Now()
		log.Printf(
			"%v ops in %v (%v/op)",
			i,
			t1.Sub(t0),
			t1.Sub(t0)/time.Duration(i),
		)
	}
	return n, true
}

type taps map[int]tapFunc
type tapFunc func(n, ip int, vm regs)

func dumpTap(n, ip int, vm regs) { log.Printf("n=%v ip=%v %v", n, ip, vm) }

var firstReg3N = make(map[int]int, 1024)

func maxReg3Tap(n, ip int, vm regs) {
	r3 := vm[3]
	if _, def := firstReg3N[r3]; !def {
		firstReg3N[r3] = n
		log.Printf("best so far: n=%v ip=%v r3=%v", n, ip, r3)
	}
}

var tapActs = map[string]tapFunc{
	"dump":   dumpTap,
	"max_r3": maxReg3Tap,
}

func (ts taps) String() string {
	return fmt.Sprint(map[int]tapFunc(ts))
}

var tapArgPattern = regexp.MustCompile(`^(\d+)(?::(\w+))?$`)

func (ts taps) Set(arg string) error {
	parts := tapArgPattern.FindStringSubmatch(arg)
	if len(parts) == 0 {
		return fmt.Errorf("invalid tap arg %q, expected %v", arg, tapArgPattern)
	}

	tapFn := dumpTap

	if parts[2] != "" {
		act := parts[2]
		if tapFn = tapActs[act]; tapFn == nil {
			return fmt.Errorf("invalid tap action %q", act)
		}
	}

	ip, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	ts[ip] = tapFn
	return nil
}

func run(in, out *os.File) error {
	ipReg, prog, err := read(in)
	if err != nil {
		return err
	}

	vm := regs{}
	vm[0] = *initR0
	vm[1] = *initR1
	vm[2] = *initR2
	vm[3] = *initR3
	vm[4] = *initR4
	vm[5] = *initR5

	if n, ok := prog.run(ipReg, vm, *limit); !ok {
		log.Printf("NOPE")
	} else {
		log.Printf("OK after %v", n)
	}

	return nil
}

var (
	ipPat = regexp.MustCompile(`#ip (\d+)`)
	opPat = regexp.MustCompile(`(\w+) (\d+) (\d+) (\d+)`)
)

func read(r io.Reader) (ipReg int, prog program, err error) {
	sc := bufio.NewScanner(r)
	if sc.Scan() {
		line := sc.Text()
		parts := ipPat.FindStringSubmatch(line)
		if len(parts) == 0 {
			return 0, nil, fmt.Errorf("unexpected line %q expected %v", line, ipPat)
		}
		ipReg, _ = strconv.Atoi(parts[1])
	}
	for sc.Scan() {
		line := sc.Text()
		parts := opPat.FindStringSubmatch(line)
		if len(parts) == 0 {
			return 0, nil, fmt.Errorf("unexpected line %q expected %v", line, opPat)
		}
		code, def := name2opCode[parts[1]]
		if !def {
			return 0, nil, fmt.Errorf("invalid op %q", parts[1])
		}

		var in inst
		in[0] = code
		in[1], _ = strconv.Atoi(parts[2])
		in[2], _ = strconv.Atoi(parts[3])
		in[3], _ = strconv.Atoi(parts[4])
		prog = append(prog, in)
	}

	return ipReg, prog, sc.Err()
}

package main

import (
	"aoc/internal/elvm"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jcorbin/anansi"
)

var tap = taps{}

var (
	dis     = flag.Bool("dis", false, "dump descriptive text")
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

type taps map[int]tapFunc
type tapFunc func(vm *elvm.VM, in elvm.Instruction) error

func dumpTap(vm *elvm.VM, in elvm.Instruction) error {
	log.Print(vm)
	return nil
}

// TODO refactor this into a proper "comparatorSpy"
var (
	firstReg3N = make(map[int]int, 1024)
	lastReg3T  time.Time
	lastReg3N  int
	lastReg3   int
)

func maxReg3Tap(vm *elvm.VM, in elvm.Instruction) error {
	const stopAfter = time.Second
	r3 := vm.R[3]
	if _, def := firstReg3N[r3]; !def {
		firstReg3N[r3] = vm.N
		if len(firstReg3N) == 1 {
			log.Printf("first halt: %v vm:%v", r3, vm)
		}
		lastReg3T = time.Now()
		lastReg3N = vm.N
		lastReg3 = r3
	} else if since := time.Now().Sub(lastReg3T); !lastReg3T.IsZero() && since > stopAfter {
		log.Printf("last halt: %v vm:%v", lastReg3, vm)
		return fmt.Errorf("stopping since we haven't seen a better R3 in the last %v (%v ops)", stopAfter, vm.N-lastReg3N)
	}
	return nil
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
	prog, err := elvm.DecodeProgram(in)
	if err != nil {
		return err
	}

	if *dis {
		w := 1
		for n := len(prog.Ops); n > 0; n /= 10 {
			w++
		}
		for ip := range prog.Ops {
			fmt.Printf("% *d %s\n", w, ip, prog.Describe(ip))
		}
		return nil
	}

	var vm elvm.VM
	if err := vm.Load(prog); err != nil {
		return err
	}

	vm.R[0] = *initR0
	vm.R[1] = *initR1
	vm.R[2] = *initR2
	vm.R[3] = *initR3
	vm.R[4] = *initR4
	vm.R[5] = *initR5

	var trc elvm.Tracer

	if *verbose {
		bw := bufio.NewWriterSize(os.Stdout, 64*1024)
		trc = elvm.Tracers(trc, &elvm.PrintTracer{W: bw})
		defer bw.Flush()
	}

	if len(tap) > 0 {
		trc = elvm.Tracers(trc, tap)
	}

	// TODO run the VM in chunks when not limited, listening for SIGINT
	// and stopping gracefully.
	defer func(n0 int, t0 time.Time) {
		n1, t1 := vm.N, time.Now()
		en, et := n1-n0, t1.Sub(t0)
		log.Printf("%v ops in %v (%v/op)", en, et, et/time.Duration(en))
	}(vm.N, time.Now())
	return vm.Execute(*limit, trc)
}

func (ts taps) Before(vm *elvm.VM, in elvm.Instruction) error {
	if tapFn := ts[*vm.IP]; tapFn != nil {
		if err := tapFn(vm, in); err != nil {
			return err
		}
	}
	return nil
}

func (ts taps) After(vm *elvm.VM, in elvm.Instruction) error {
	return nil
}

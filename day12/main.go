package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jcorbin/anansi"
)

var (
	numGenerations = flag.Int("n", 20, "number of generations to run")
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type space struct {
	min, max int
	pots     map[int]bool // TODO is a bitvector worth it?
	rules    map[ruleKey]bool
}

type ruleKey uint8

const (
	ruleL2 ruleKey = 1 << iota
	ruleL1
	ruleC
	ruleR1
	ruleR2
)

func rule(l2, l1, c, r1, r2 bool) (r ruleKey) {
	if l2 {
		r |= ruleL2
	}
	if l1 {
		r |= ruleL1
	}
	if c {
		r |= ruleC
	}
	if r1 {
		r |= ruleR1
	}
	if r2 {
		r |= ruleR2
	}
	return r
}

func parseRuleParts(s string) (l2, l1, c, r1, r2 bool) {
	if len(s) > 0 {
		l2 = s[0] == '#'
	}
	if len(s) > 1 {
		l1 = s[1] == '#'
	}
	if len(s) > 2 {
		c = s[2] == '#'
	}
	if len(s) > 3 {
		r1 = s[3] == '#'
	}
	if len(s) > 4 {
		r2 = s[4] == '#'
	}
	return l2, l1, c, r1, r2
}

func run(in, out *os.File) error {
	spc, err := read(in)
	if err != nil {
		return err
	}

	// part 1
	for i := 0; i < *numGenerations; i++ {
		spc.tick()
	}

	var buf bytes.Buffer

	for i := spc.min; i < spc.max; i++ {
		if i != 0 && i%10 == 0 {
			buf.WriteByte('0' + byte(i/10%10))
		} else if i < 0 && i%10 == 9 {
			buf.WriteByte('-')
		} else {
			buf.WriteByte(' ')
		}
	}
	buf.WriteByte('\n')
	for i := spc.min; i < spc.max; i++ {
		if i%10 == 0 {
			buf.WriteByte('0')
		} else {
			buf.WriteByte(' ')
		}
	}
	buf.WriteByte('\n')
	buf.WriteTo(os.Stdout)

	for i := spc.min; i < spc.max; i++ {
		if spc.pots[i] {
			buf.WriteByte('#')
		} else {
			buf.WriteByte('.')
		}
	}
	buf.WriteByte('\n')
	buf.WriteTo(os.Stdout)

	log.Printf("min:%v max:%v", spc.min, spc.max)
	n := 0
	for i := spc.min; i < spc.max; i++ {
		if spc.pots[i] {
			n += i
		}
	}

	log.Printf("total: %v", n)

	return nil
}

func (spc *space) tick() {
	var vs [5]bool
	load := spc.min
	stor := spc.min - 2
	max := spc.max + 5
	for stor < max {
		vs[0], vs[1], vs[2], vs[3], vs[4] = vs[1], vs[2], vs[3], vs[4], spc.pots[load]
		v := spc.rules[rule(vs[0], vs[1], vs[2], vs[3], vs[4])]
		if in := spc.min <= stor && stor < spc.max; v || in {
			spc.pots[stor] = v
			if spc.min > stor {
				spc.min = stor
			}
			if spc.max <= stor {
				spc.max = stor + 1
			}
		}
		load++
		stor++
	}
}

func read(r io.Reader) (spc space, _ error) {
	sc := bufio.NewScanner(r)
	if sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "initial state: ") {
			return spc, errors.New("no initial sate line")
		}
		init := strings.TrimPrefix(line, "initial state: ")
		spc.pots = make(map[int]bool, 2*len(init))
		spc.min = 0
		spc.max = len(init)
		for i := 0; i < len(init); i++ {
			if init[i] == '#' {
				spc.pots[i] = true
			}
		}
	}
	sc.Scan()
	spc.rules = make(map[ruleKey]bool, 64)
	for sc.Scan() {
		line := sc.Text()
		if parts := strings.SplitN(line, " => ", 2); len(parts) == 2 {
			rk := rule(parseRuleParts(parts[0]))
			spc.rules[rk] = parts[1][0] == '#'
		}
	}
	return spc, sc.Err()
}

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

const chunkSize = 64

type space struct {
	rules  [256]bool
	offset int
	chunks []uint64
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

	for i := 0; i < *numGenerations; i++ {
		spc.tick()
	}

	var buf bytes.Buffer

	// for i := spc.min; i < spc.max; i++ {
	// 	if i != 0 && i%10 == 0 {
	// 		buf.WriteByte('0' + byte(i/10%10))
	// 	} else if i < 0 && i%10 == 9 {
	// 		buf.WriteByte('-')
	// 	} else {
	// 		buf.WriteByte(' ')
	// 	}
	// }
	// buf.WriteByte('\n')
	// for i := spc.min; i < spc.max; i++ {
	// 	if i%10 == 0 {
	// 		buf.WriteByte('0')
	// 	} else {
	// 		buf.WriteByte(' ')
	// 	}
	// }
	// buf.WriteByte('\n')
	// buf.WriteTo(os.Stdout)

	spc.writePotBytes(&buf)
	buf.WriteByte('\n')
	buf.WriteTo(os.Stdout)

	log.Printf("total: %v", spc.sumPots())

	return nil
}

func (spc *space) min() int {
	for i, c := range spc.chunks {
		m := uint64(1)
		for j := 0; j < chunkSize; j++ {
			if c&m != 0 {
				return spc.offset + i*chunkSize + j
			}
			m <<= 1
		}
	}
	return 0
}

func (spc *space) max() int {
	var max int
	for i, c := range spc.chunks {
		m := uint64(1)
		for j := 0; j < chunkSize; j++ {
			if c&m != 0 {
				max = spc.offset + i*chunkSize + j + 1
			}
			m <<= 1
		}
	}
	return max
}

func (spc *space) writePotBytes(buf *bytes.Buffer) {
	started := false
	skip := 0
	for _, c := range spc.chunks {
		m := uint64(1)
		for j := 0; j < chunkSize; j++ {
			if c&m != 0 {
				if started {
					for k := 0; k < skip; k++ {
						buf.WriteByte('.')
					}
				}
				skip = 0
				buf.WriteByte('#')
				started = true
			} else {
				skip++
			}
			m <<= 1
		}
	}
}

func (spc *space) sumPots() (n int) {
	for i, c := range spc.chunks {
		m := uint64(1)
		for j := 0; j < chunkSize; j++ {
			if c&m != 0 {
				n += spc.offset + i*chunkSize + j
			}
			m <<= 1
		}
	}
	return n
}

func (spc *space) tick() {
	var nl, nr uint64
	var vs [5]bool
	out := -2
	i := 0
	for ; i < len(spc.chunks); i++ {
		c := spc.chunks[i]
		nc := uint64(0)
		m := uint64(1)
		for j := 0; j < chunkSize; j++ {
			vs[0], vs[1], vs[2], vs[3], vs[4] = vs[1], vs[2], vs[3], vs[4], c&m != 0
			rv := spc.rules[rule(vs[0], vs[1], vs[2], vs[3], vs[4])]
			if rv {
				if out >= 0 {
					nc |= 1 << uint64(out)
				} else if i == 0 {
					nl |= 1 << uint64(chunkSize+out)
				} else {
					spc.chunks[i-1] |= 1 << uint64(chunkSize+out)
				}
			}
			m <<= 1
			out++
		}
		spc.chunks[i] = nc
		out = -2
	}

	vs[0], vs[1], vs[2], vs[3], vs[4] = vs[1], vs[2], vs[3], vs[4], false
	if spc.rules[rule(vs[0], vs[1], vs[2], vs[3], vs[4])] {
		spc.chunks[i-1] |= 1 << uint64(chunkSize+out)
	}
	out++

	vs[0], vs[1], vs[2], vs[3], vs[4] = vs[1], vs[2], vs[3], vs[4], false
	if spc.rules[rule(vs[0], vs[1], vs[2], vs[3], vs[4])] {
		spc.chunks[i-1] |= 1 << uint64(chunkSize+out)
	}
	out++

	vs[0], vs[1], vs[2], vs[3], vs[4] = vs[1], vs[2], vs[3], vs[4], false
	if spc.rules[rule(vs[0], vs[1], vs[2], vs[3], vs[4])] {
		nr |= 1
	}

	vs[0], vs[1], vs[2], vs[3], vs[4] = vs[1], vs[2], vs[3], vs[4], false
	if spc.rules[rule(vs[0], vs[1], vs[2], vs[3], vs[4])] {
		nr |= 2
	}

	if nl != 0 {
		spc.chunks = append([]uint64{nl}, spc.chunks...)
		spc.offset -= chunkSize
	}
	if nr != 0 {
		spc.chunks = append(spc.chunks, nr)
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

		c := uint64(0)
		i := 0
		for ; i < len(init); i++ {
			if i > 0 && i%chunkSize == 0 {
				spc.chunks = append(spc.chunks, c)
				c = 0
			}
			if init[i] == '#' {
				c |= 1 << uint64(i%64)
			}
		}
		if i > 0 && i%chunkSize != 0 {
			spc.chunks = append(spc.chunks, c)
		}

	}
	sc.Scan()
	for sc.Scan() {
		line := sc.Text()
		if parts := strings.SplitN(line, " => ", 2); len(parts) == 2 {
			rk := rule(parseRuleParts(parts[0]))
			spc.rules[rk] = parts[1][0] == '#'
		}
	}
	return spc, sc.Err()
}

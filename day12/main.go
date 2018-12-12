package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	"github.com/jcorbin/anansi"
)

var (
	numGenerations = flag.Int("n", 20, "number of generations to run")
	bitmap         = flag.Bool("b", false, "output a compact bitmap")
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
	ruleR2 ruleKey = 1 << iota
	ruleR1
	ruleC
	ruleL1
	ruleL2
)

const ruleMask = ruleR2 | ruleR1 | ruleC | ruleL1 | ruleL2

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

type sim struct {
	spc  space
	tick int

	min, max int

	traceBuf   bytes.Buffer
	mins, maxs []int
	total      []int
	byteLens   []int
}

func (sim *sim) step() {
	sim.spc.tick()
	sim.tick++
	sim.snapshot()
}

func (sim *sim) snapshot() {
	min, max := sim.spc.min(), sim.spc.max()
	if sim.min > min {
		sim.min = min
	}
	if sim.max < max {
		sim.max = max
	}
	sim.mins = append(sim.mins, min)
	sim.maxs = append(sim.maxs, max)
	sim.total = append(sim.total, sim.spc.sumPots())
	l0 := sim.traceBuf.Len()
	sim.spc.writePotBytes(&sim.traceBuf)
	l1 := sim.traceBuf.Len()
	sim.byteLens = append(sim.byteLens, l1-l0)
}

func (sim *sim) dumpTraceTo(w io.Writer) error {
	var buf bytes.Buffer

	// compute T and Σ padding
	iw := tens(len(sim.byteLens) - 1)
	tw := 0
	for _, t := range sim.total {
		if n := tens(t); tw < n {
			tw = n
		}
	}

	// header 1
	fmt.Fprintf(&buf, "% *s % *s", iw, "", tw, "")
	for i := sim.min; i < sim.max; i++ {
		if i != 0 && i%10 == 0 {
			buf.WriteByte('0' + byte(i/10%10))
		} else if i < 0 && i%10 == 9 {
			buf.WriteByte('-')
		} else {
			buf.WriteByte(' ')
		}
	}
	buf.WriteByte('\n')

	// header 2
	fmt.Fprintf(&buf, "% *s % *s", iw, "T", tw, "Σ")
	for i := sim.min; i < sim.max; i++ {
		if i%10 == 0 {
			buf.WriteByte('0')
		} else {
			buf.WriteByte(' ')
		}
	}
	buf.WriteByte('\n')
	if _, err := buf.WriteTo(w); err != nil {
		return err
	}

	// padded rows
	byteOff := 0
	for i, n := range sim.byteLens {
		min, max := sim.mins[i], sim.maxs[i]
		byteEnd := byteOff + n
		fmt.Fprintf(&buf, "% *d % *d ", iw, i, tw, sim.total[i])
		for i := sim.min; i < min; i++ {
			buf.WriteByte('.')
		}
		buf.Write(sim.traceBuf.Bytes()[byteOff:byteEnd])
		for i := max; i < sim.max; i++ {
			buf.WriteByte('.')
		}
		buf.WriteByte('\n')
		byteOff = byteEnd
		if _, err := buf.WriteTo(w); err != nil {
			return err
		}
	}

	return nil
}

func (sim *sim) makeBitmap() (bi anansi.Bitmap) {
	bi.Rect.Max.Y = sim.tick
	bi.Rect.Max.X = sim.max - sim.min
	bi.Stride = bi.Rect.Max.X
	bi.Bit = make([]bool, bi.Stride*bi.Rect.Dy())
	byteOff := 0
	for i, n := range sim.byteLens {
		byteEnd := byteOff + n
		off := sim.mins[i] - sim.min
		line := sim.traceBuf.Bytes()[byteOff:byteEnd]
		for x, c := range line {
			bi.Set(image.Pt(off+x, i), c == '#')
		}
		byteOff = byteEnd
	}
	return bi
}

func run(in, out *os.File) error {
	spc, err := read(in)
	if err != nil {
		return err
	}

	sim := sim{spc: spc}
	sim.snapshot()
	for sim.tick < *numGenerations {
		sim.step()
	}

	if *bitmap {
		bi := sim.makeBitmap()
		_, err = anansi.WriteBitmap(os.Stdout, &bi)
		return err
	}

	return sim.dumpTraceTo(os.Stdout)
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

func (rk ruleKey) lshift(bit ruleKey) ruleKey {
	rk = (rk<<1)&ruleMask | bit
	return rk
}

func (spc *space) tick() {
	var nl, nr uint64
	var rk ruleKey
	out := -2
	i := 0
	for ; i < len(spc.chunks); i++ {
		c := spc.chunks[i]
		nc := uint64(0)
		for j := 0; j < chunkSize; j++ {
			rk = rk.lshift(ruleKey(c & 1))
			if spc.rules[rk] {
				if out >= 0 {
					nc |= 1 << uint64(out)
				} else if i == 0 {
					nl |= 1 << uint64(chunkSize+out)
				} else {
					spc.chunks[i-1] |= 1 << uint64(chunkSize+out)
				}
			}
			c >>= 1
			out++
		}
		spc.chunks[i] = nc
		out = -2
	}

	rk = rk.lshift(0)
	if spc.rules[rk] {
		spc.chunks[i-1] |= 1 << uint64(chunkSize+out)
	}
	out++

	rk = rk.lshift(0)
	if spc.rules[rk] {
		spc.chunks[i-1] |= 1 << uint64(chunkSize+out)
	}
	out++

	rk = rk.lshift(0)
	if spc.rules[rk] {
		nr |= 1
	}

	rk = rk.lshift(0)
	if spc.rules[rk] {
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

func tens(n int) (m int) {
	for n > 0 {
		m++
		n /= 10
	}
	return m + 1
}

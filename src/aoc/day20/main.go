package main

import (
	"aoc/internal/progprof"
	"aoc/internal/quadindex"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"

	"github.com/jcorbin/anansi"
)

var (
	errInputStart = errors.New("invalid room pattern: must begin with ^")
	errInputEnd   = errors.New("invalid input, doesn't end in $")
)

var (
	drawFlag = flag.Bool("draw", false, "draw rather than solve")
	patFlag  = flag.String("pat", "", "pattern from cli rather than stdin")
)

func main() {
	flag.Parse()
	defer progprof.Start()()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type pointScore map[image.Point]int
type pointGraph map[image.Point]pointSet
type pointSet map[image.Point]struct{}

func run(in, out *os.File) error {
	pattern := *patFlag

	if pattern == "" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		log.Printf("read a %v-byte pattern from stdin", len(b))
		for i := len(b) - 1; b[i] == '\n'; {
			b = b[:i]
			i--
		}
		pattern = string(b)
	}

	// part 1
	log.Printf("building rooms from pattern")
	var bld builder
	if err := bld.buildRooms(pattern); err != nil {
		return err
	}

	if *drawFlag {
		log.Printf("drawing rooms")
		var rm roomMap
		bld.pg.draw(&rm, image.ZP, nil)
		fmt.Printf("%s\n", rm.draw())
	}

	log.Printf("filling rooms")
	// XXX why doesn't this return the right max value?
	pd := make(pointScore, len(bld.pg))
	log.Printf("farthest: %v", bld.pg.fill(image.ZP, 0, pd))

	n := 0
	for _, d := range pd {
		if d >= 1000 {
			n++
		}
	}
	log.Printf("1000 or over: %v", n)

	if *drawFlag {
		log.Printf("drawing rooms (again)")
		var rm roomMap
		bld.pg.draw(&rm, image.ZP, pd)
		fmt.Printf("%s\n", rm.draw())
	}

	return nil
}

func (pg pointGraph) connected(a, b image.Point) bool {
	_, con := pg[a][b]
	return con
}

func (pg pointGraph) fill(p image.Point, d int, pd pointScore) int {
	if rv, def := pd[p]; def {
		return rv
	}

	pd[p] = d
	rv := d
	for _, np := range [...]image.Point{
		p.Add(image.Pt(0, -1)),
		p.Add(image.Pt(1, 0)),
		p.Add(image.Pt(0, 1)),
		p.Add(image.Pt(-1, 0)),
	} {
		if pg.connected(p, np) {
			rv = max(rv, pg.fill(np, d+1, pd))
		}
	}
	return rv
}

func max(a int, bs ...int) int {
	for _, b := range bs {
		if a < b {
			a = b
		}
	}
	return a
}

type builder struct {
	pg pointGraph

	// TODO surely this is too simple
	st []image.Point
}

func (bld *builder) buildRooms(pattern string) error {
	bld.pg = make(pointGraph, 1024)
	bld.st = make([]image.Point, 0, 1024)

	i := 0
	if pattern[i] != '^' {
		return errInputStart
	}
	i++

	cur := image.ZP
	for ; i < len(pattern)-1; i++ {
		switch pattern[i] {
		case 'N':
			cur = bld.add(cur, cur.Add(image.Pt(0, -1)))
		case 'E':
			cur = bld.add(cur, cur.Add(image.Pt(1, 0)))
		case 'S':
			cur = bld.add(cur, cur.Add(image.Pt(0, 1)))
		case 'W':
			cur = bld.add(cur, cur.Add(image.Pt(-1, 0)))
		case '(':
			bld.st = append(bld.st, cur)
		case '|':
			cur = bld.st[len(bld.st)-1]
		case ')':
			bld.st = bld.st[:len(bld.st)-1]
		default:
			return fmt.Errorf("invalid pattern input %q", string(pattern[i]))
		}
	}

	if pattern[i] != '$' {
		return errInputEnd
	}
	return nil
}

func (bld *builder) add(a, b image.Point) image.Point {
	aSet := bld.pg[a]
	bSet := bld.pg[b]
	if aSet == nil {
		aSet = make(pointSet, 64)
		bld.pg[a] = aSet
	}
	if bSet == nil {
		bSet = make(pointSet, 64)
		bld.pg[b] = bSet
	}
	aSet[b] = struct{}{}
	bSet[a] = struct{}{}
	return b
}

type roomMap struct {
	bounds image.Rectangle

	quadindex.Index
	p []image.Point
	r []byte
}

func (rm *roomMap) draw() string {
	var buf bytes.Buffer
	for p := rm.bounds.Min; p.Y < rm.bounds.Max.Y; p.Y++ {
		if p.Y > rm.bounds.Min.Y {
			buf.WriteByte('\n')
		}
		for p.X = rm.bounds.Min.X; p.X < rm.bounds.Max.X; p.X++ {
			if cur := rm.At(p); cur.Next() {
				id := cur.I()
				buf.WriteByte(rm.r[id])
			} else {
				buf.WriteByte(' ')
			}
		}
	}
	return buf.String()
}

func (rm *roomMap) setCell(p image.Point, r byte) bool {
	if cur := rm.Index.At(p); cur.Next() {
		id := cur.I()
		if rm.r[id] == r {
			return false
		}
		rm.r[id] = r
		return true
	}

	id := len(rm.p)
	rm.p = append(rm.p, p)
	rm.r = append(rm.r, r)
	rm.Index.Update(id, p)

	pr := image.Rectangle{p, p.Add(image.Pt(1, 1))}
	if rm.bounds == image.ZR {
		rm.bounds = pr
	} else {
		rm.bounds = rm.bounds.Union(pr)
	}

	return true
}

func (pg pointGraph) draw(rm *roomMap, p image.Point, pd pointScore) {
	for _, corner := range [...]image.Point{
		image.Pt(-1, -1),
		image.Pt(1, -1),
		image.Pt(-1, 1),
		image.Pt(1, 1),
	} {
		rm.setCell(p.Mul(2).Add(corner), '#')
	}

	for _, door := range [...]struct {
		d image.Point
		r byte
	}{
		{image.Pt(0, -1), '-'},
		{image.Pt(-1, 0), '|'},
		{image.Pt(0, 1), '-'},
		{image.Pt(1, 0), '|'},
	} {
		if np := p.Add(door.d); !pg.connected(p, np) {
			rm.setCell(p.Mul(2).Add(door.d), '#')
		} else if rm.setCell(p.Mul(2).Add(door.d), door.r) {
			pg.draw(rm, np, pd)
		}
	}

	if p == image.ZP {
		rm.setCell(p.Mul(2), 'X')
	} else if d, def := pd[p]; !def {
		rm.setCell(p.Mul(2), '.')
	} else {
		if d < 10 {
			rm.setCell(p.Mul(2), '0'+byte(d))
		} else if d < 36 {
			rm.setCell(p.Mul(2), 'a'+byte(d-10))
		} else if d < 62 {
			rm.setCell(p.Mul(2), 'A'+byte(d-36))
		} else if d < 62 {
			rm.setCell(p.Mul(2), '?')
		}
	}

}

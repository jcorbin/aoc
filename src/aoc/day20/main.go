package main

import (
	"aoc/internal/quadindex"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/jcorbin/anansi"
)

var (
	errInputStart = errors.New("invalid room pattern: must begin with ^")
	errInputEnd   = errors.New("invalid input, doesn't end in $")
)

var (
	drawFlag = flag.Bool("draw", false, "draw rather than solve")
	patFlag  = flag.String("pat", "", "pattern from cli rather than stdin")

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		log.Printf("CPU profiling to %q", f.Name())
		defer pprof.StopCPUProfile()
	}

	if *cpuprofile != "" || *memprofile != "" {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGINT)
			<-ch
			signal.Stop(ch)
			if *cpuprofile != "" {
				pprof.StopCPUProfile()
				log.Printf("stopped CPU profiling")
			}
			takeMemProfile()
		}()
	}

	anansi.MustRun(run(os.Stdin, os.Stdout))

	takeMemProfile()
}

func takeMemProfile() {
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		log.Printf("heap profile to %q", f.Name())
		f.Close()
	}
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

	// if *drawFlag {
	// 	log.Printf("drawing rooms")
	// 	var rm roomMap
	// 	start.build(&rm, image.ZP)
	// 	fmt.Printf("%s\n", rm.draw())
	// }

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

	// if *drawFlag {
	// 	log.Printf("drawing rooms (again)")
	// 	var rm roomMap
	// 	start.build(&rm, image.ZP)
	// 	fmt.Printf("%s\n", rm.draw())
	// }

	return nil
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
		if _, con := pg[p][np]; con {
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

// func (bld *builder) draw(rm *roomMap, p image.Point) {
// 	rm.setCell(p.Add(image.Pt(-1, -1)), '#')
// 	if pp := p.Add(image.Pt(0, -1)); r.n == nil {
// 		rm.setCell(pp, '#')
// 	} else if rm.setCell(pp, '-') {
// 		r.n.build(rm, pp.Add(image.Pt(0, -1)))
// 	}
// 	rm.setCell(p.Add(image.Pt(1, -1)), '#')

// 	if pp := p.Add(image.Pt(-1, 0)); r.w == nil {
// 		rm.setCell(pp, '#')
// 	} else if rm.setCell(pp, '|') {
// 		r.w.build(rm, pp.Add(image.Pt(-1, 0)))
// 	}

// 	if p == image.ZP {
// 		rm.setCell(p, 'X')
// 	} else if r.d < 0 {
// 		rm.setCell(p, '.')
// 	} else {
// 		if r.d < 10 {
// 			rm.setCell(p, '0'+byte(r.d))
// 		} else if r.d < 36 {
// 			rm.setCell(p, 'a'+byte(r.d-10))
// 		} else if r.d < 62 {
// 			rm.setCell(p, 'A'+byte(r.d-36))
// 		} else if r.d < 62 {
// 			rm.setCell(p, '?')
// 		}
// 	}

// 	if pp := p.Add(image.Pt(1, 0)); r.e == nil {
// 		rm.setCell(pp, '#')
// 	} else if rm.setCell(pp, '|') {
// 		r.e.build(rm, pp.Add(image.Pt(1, 0)))
// 	}

// 	rm.setCell(p.Add(image.Pt(-1, 1)), '#')
// 	if pp := p.Add(image.Pt(0, 1)); r.s == nil {
// 		rm.setCell(pp, '#')
// 	} else if rm.setCell(pp, '-') {
// 		r.s.build(rm, pp.Add(image.Pt(0, 1)))
// 	}
// 	rm.setCell(p.Add(image.Pt(1, 1)), '#')
// }

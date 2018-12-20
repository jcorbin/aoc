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
	"time"

	"github.com/jcorbin/anansi"
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

var hack int // FIXME

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
	start, err := bld.buildRooms(pattern)
	if err != nil {
		return err
	}

	if *drawFlag {
		log.Printf("drawing rooms")
		var rm roomMap
		start.build(&rm, image.ZP)
		fmt.Printf("%s\n", rm.draw())
	}

	log.Printf("filling rooms")
	// XXX why doesn't this return the right max value?
	log.Printf("%v", start.fill(0)+1)

	hack = 0
	if *drawFlag {
		log.Printf("drawing rooms (again)")
		var rm roomMap
		start.build(&rm, image.ZP)
		fmt.Printf("%s\n", rm.draw())
	}
	log.Printf("%v", hack)

	// part 2
	// TODO

	return nil
}

func (r *room) fill(d int) int {
	if r == nil {
		return -1
	}
	// TODO is this necessary?
	if r.d >= 0 && r.d <= d {
		return r.d
	}
	r.d = d
	n, e, s, w := r.n.fill(d+1), r.e.fill(d+1), r.s.fill(d+1), r.w.fill(d+1)
	// log.Printf("FILL %v %v %v %v", n, e, s, w) XXX why ...
	return max(0, n, e, s, w)
}

func max(a int, bs ...int) int {
	for _, b := range bs {
		if a < b {
			a = b
		}
	}
	return a
}

type room struct {
	p          image.Point
	n, e, s, w *room
	d          int
}

type buildState struct {
	*builder

	i    int
	cur  *room
	pend []*room

	curStack []*room
}

var (
	slab      []room
	slabi     int
	allocated int
)

func (bs *buildState) grow(p image.Point) *room {
	if cur := bs.At(p); cur.Next() {
		bs.cur = bs.rooms[cur.I()]
	} else {
		if slabi >= len(slab) {
			slab = make([]room, 1024*1024)
			slabi = 0
			allocated += len(slab)
		}
		bs.cur = &slab[slabi]
		slabi++
		bs.cur.p = p
		bs.cur.d = -1

		i := len(bs.rooms)
		bs.rooms = append(bs.rooms, bs.cur)
		bs.Index.Update(i, p)
	}
	return bs.cur
}

func (bs *buildState) push() {
	bs.curStack = append(bs.curStack, bs.cur)
	bs.pend = nil
}

func (bs *buildState) pop(emit func(buildState)) {
	csi := len(bs.curStack) - 1
	for _, r := range bs.pend {
		i := bs.i + 1

		if _, seen := bs.seen[i][r]; seen {
			continue
		}
		seenRooms := bs.seen[i]
		if seenRooms == nil {
			if bs.seen == nil {
				bs.seen = make(map[int]map[*room]struct{}, 64)
			}
			seenRooms = make(map[*room]struct{}, 64)
			bs.seen[i] = seenRooms
		}
		seenRooms[r] = struct{}{}

		next := buildState{
			builder: bs.builder,
			i:       i,
			cur:     r,
		}
		if csi > 0 {
			next.curStack = bs.curStack[:csi:csi]
		}
		emit(next)
	}
	bs.curStack = bs.curStack[:csi]
}

func (bs *buildState) alt() {
	bs.pend = append(bs.pend, bs.cur)
	bs.cur = bs.curStack[len(bs.curStack)-1]
}

var (
	errInputStart = errors.New("invalid room pattern: must begin with ^")
	errInputEnd   = errors.New("invalid input, doesn't end in $")
)

func (bs *buildState) expand(pattern string, emit func(buildState)) error {
	// TODO rework this to use a set of currently expanding points rather than
	// backtracking:
	// - start opens a new context (stack push prior)
	// - alt adds a continuation to the next gen
	// - end takes the current set of next continuations and runs with them
	for ; bs.i < len(pattern); bs.i++ {
		prior := bs.cur
		// log.Printf("> %q in %p", string(pattern[bs.i]), bs.cur)
		switch pattern[bs.i] {
		case 'N':
			prior.n = bs.grow(prior.p.Add(image.Pt(0, -1)))
			bs.cur.s = prior
		case 'E':
			prior.e = bs.grow(prior.p.Add(image.Pt(1, 0)))
			bs.cur.w = prior
		case 'S':
			prior.s = bs.grow(prior.p.Add(image.Pt(0, 1)))
			bs.cur.n = prior
		case 'W':
			prior.w = bs.grow(prior.p.Add(image.Pt(-1, 0)))
			bs.cur.e = prior
		case '(':
			bs.push()
		case '|':
			bs.alt()
		case ')':
			bs.pop(emit)
		case '$':
			bs.i++
			if bs.i == len(pattern) {
				return nil
			}
			bs.i = len(pattern)

		default:
			return fmt.Errorf("invalid pattern input %q", string(pattern[bs.i]))
		}
		// log.Printf("... %p", bs.cur)
	}

	return errInputEnd
}

type builder struct {
	quadindex.Index
	rooms []*room
	seen  map[int]map[*room]struct{}
}

func (bld *builder) buildRooms(pattern string) (*room, error) {
	var start buildState

	if pattern[start.i] != '^' {
		return nil, errInputStart
	}
	start.builder = bld
	start.grow(image.ZP)
	start.i++

	frontier := []buildState{start}
	emit := func(alt buildState) {
		// log.Printf("EMIT %q ^ %q", pattern[:alt.i], pattern[alt.i:])
		frontier = append(frontier, alt)
	}

	tick := time.NewTicker(time.Second)
	n := 0
	for len(frontier) > 0 {
		n++
		st := frontier[len(frontier)-1]
		frontier = frontier[:len(frontier)-1]

		select {
		case <-tick.C:
			log.Printf("expanded %v states, depth %v @%v allocated:%v", n, len(frontier), st.i, allocated)
		default:
		}

		if err := st.expand(pattern, emit); err != nil {
			return nil, err
		}
	}

	return start.cur, nil
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

func (r *room) build(rm *roomMap, p image.Point) {
	rm.setCell(p.Add(image.Pt(-1, -1)), '#')
	if pp := p.Add(image.Pt(0, -1)); r.n == nil {
		rm.setCell(pp, '#')
	} else if rm.setCell(pp, '-') {
		r.n.build(rm, pp.Add(image.Pt(0, -1)))
	}
	rm.setCell(p.Add(image.Pt(1, -1)), '#')

	if pp := p.Add(image.Pt(-1, 0)); r.w == nil {
		rm.setCell(pp, '#')
	} else if rm.setCell(pp, '|') {
		r.w.build(rm, pp.Add(image.Pt(-1, 0)))
	}

	if p == image.ZP {
		rm.setCell(p, 'X')
	} else if r.d < 0 {
		rm.setCell(p, '.')
	} else {
		if hack < r.d {
			hack = r.d
		}
		if r.d < 10 {
			rm.setCell(p, '0'+byte(r.d))
		} else if r.d < 36 {
			rm.setCell(p, 'a'+byte(r.d-10))
		} else if r.d < 62 {
			rm.setCell(p, 'A'+byte(r.d-36))
		} else if r.d < 62 {
			rm.setCell(p, '?')
		}
	}

	if pp := p.Add(image.Pt(1, 0)); r.e == nil {
		rm.setCell(pp, '#')
	} else if rm.setCell(pp, '|') {
		r.e.build(rm, pp.Add(image.Pt(1, 0)))
	}

	rm.setCell(p.Add(image.Pt(-1, 1)), '#')
	if pp := p.Add(image.Pt(0, 1)); r.s == nil {
		rm.setCell(pp, '#')
	} else if rm.setCell(pp, '-') {
		r.s.build(rm, pp.Add(image.Pt(0, 1)))
	}
	rm.setCell(p.Add(image.Pt(1, 1)), '#')
}

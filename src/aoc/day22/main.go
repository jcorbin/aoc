package main

import (
	"container/heap"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
)

// TODO re-use w/ day20
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

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

type scan struct {
	depth   int
	target  image.Point
	giCache map[image.Point]int
}

const (
	xFactor    = 16807
	yFactor    = 48271
	erosionMod = 20183
)

func (sc *scan) geologicIndex(p image.Point) (gi int) {

	if gi, def := sc.giCache[p]; def {
		return gi
	}
	defer func() {
		if sc.giCache == nil {
			sc.giCache = make(map[image.Point]int, 1024)
		}
		sc.giCache[p] = gi
	}()

	// The region at 0,0 (the mouth of the cave) has a geologic index of 0.
	if p == image.ZP {
		return 0
	}

	// The region at the coordinates of the target has a geologic index of 0.
	if p == sc.target {
		return 0
	}

	// If the region's Y coordinate is 0, the geologic index is its X coordinate times 16807.
	if p.Y == 0 {
		return (p.X * xFactor) % erosionMod
	}

	// If the region's X coordinate is 0, the geologic index is its Y coordinate times 48271.
	if p.X == 0 {
		return (p.Y * yFactor) % erosionMod
	}

	// Otherwise, the region's geologic index is the result of multiplying the erosion levels of the regions at X-1,Y and X,Y-1.
	left := sc.erosionLevel(p.Add(image.Pt(-1, 0)))
	up := sc.erosionLevel(p.Add(image.Pt(0, -1)))
	return (left * up) % erosionMod
}

func (sc *scan) erosionLevel(p image.Point) int {
	// A region's erosion level is its geologic index plus the cave system's
	// depth, all modulo 20183. Then:
	return (sc.geologicIndex(p)%erosionMod + sc.depth%erosionMod) % erosionMod
}

type regionType uint8

const (
	regionInvalid regionType = iota
	regionRocky
	regionWet
	regionNarrow
)

func (sc *scan) regionType(p image.Point) regionType {
	el := sc.erosionLevel(p)
	switch el % 3 {
	// If the erosion level modulo 3 is 0, the region's type is rocky.
	case 0:
		return regionRocky
	// If the erosion level modulo 3 is 1, the region's type is wet.
	case 1:
		return regionWet
	// If the erosion level modulo 3 is 2, the region's type is narrow.
	case 2:
		return regionNarrow
	}
	return regionInvalid
}

var (
	depthFlag = flag.Int("d", 510, "depth")
	targetX   = flag.Int("x", 10, "target X")
	targetY   = flag.Int("y", 10, "target Y")
)

type explorerState struct {
	image.Point
	torch bool
	gear  bool
}

type explorer struct {
	time   int
	id     int
	parent int
	explorerState
}

func (ex explorer) String() string {
	return fmt.Sprintf("t:%v @%v torch:%v gear:%v", ex.time, ex.Point, ex.torch, ex.gear)
}

func (ex explorer) done(sc *scan) bool {
	return ex.Point == sc.target && ex.torch
}

func (ex explorer) hasBeen(p image.Point, s *search) bool {
	if p == image.ZP {
		return true
	}
	for id := ex.id; id != 0; id = s.par[id] {
		if s.been[id] == p {
			return true
		}
	}
	return false
}

func (ex explorer) expand(s *search) explorer {
	// Finally, once you reach the target, you need the *torch* equipped before
	// you can find him in the dark. The target is always in a *rocky* region,
	// so if you arrive there with *climbing gear* equipped, you will need to
	// spend seven minutes switching to your torch.
	if ex.Point == s.target {
		if !ex.torch {
			ex.torch, ex.gear = true, false
			ex.time += 7
		}
		return ex
	}

	var retEx explorer
	for _, dp := range [...]image.Point{
		image.Pt(0, -1),
		image.Pt(1, 0),
		image.Pt(0, 1),
		image.Pt(-1, 0),
	} {
		p := ex.Point.Add(dp)
		if p.X < 0 || p.Y < 0 {
			continue
		}
		if ex.hasBeen(p, s) {
			continue
		}

		// You can move to an adjacent region (up, down, left, or right; never
		// diagonally) if your currently equipped tool allows you to enter that
		// region. Moving to an adjacent region takes one minute. (For example,
		// if you have the torch equipped, you can move between rocky and
		// narrow regions, but cannot enter wet regions.)

		// You can change your currently equipped tool or put both away if your
		// new equipment would be valid for your current region. Switching to
		// using the climbing gear, torch, or neither always takes seven
		// minutes, regardless of which tools you start with. (For example, if
		// you are in a rocky region, you can switch from the torch to the
		// climbing gear, but you cannot switch to neither.)

		next := ex.move(s, p)
		if next.time == 0 || next == ex {
			continue
		}

		if retEx.time == 0 {
			retEx = next
		} else if s.preferExplorer(next, retEx) {
			s.add(retEx)
			retEx = next
		} else {
			s.add(next)
		}

	}
	return retEx
}

func (ex explorer) move(s *search, p image.Point) explorer {
	switch s.regionType(p) {
	// In rocky regions, you can use the climbing gear or the torch.  You
	// cannot use neither (you'll likely slip and fall).
	case regionRocky:
		if ex.torch || ex.gear {
			ex.time++
			ex.Point = p
			return ex
		}

		next := ex
		next.time++
		next.Point = p
		next.time += 7
		next.torch, next.gear = true, false
		s.add(next)

		next.torch, next.gear = false, true
		s.add(next)

		next.torch, next.gear = true, true
		return next

	// In wet regions, you can use the climbing gear or neither tool.  You
	// cannot use the torch (if it gets wet, you won't have a light source).
	case regionWet:
		ex.time++
		ex.Point = p
		if ex.torch {
			ex.time += 7
			ex.torch = false
			if ex.gear {
				s.add(ex)
				ex.gear = false
			}
		} else if ex.gear {
			next := ex
			next.time += 7
			next.gear = false
			s.add(next)
		}
		return ex

	// In narrow regions, you can use the torch or neither tool. You cannot use
	// the climbing gear (it's too bulky to fit).
	case regionNarrow:
		ex.time++
		ex.Point = p
		if ex.gear {
			ex.time += 7
			ex.gear = false
			if ex.torch {
				s.add(ex)
				ex.torch = false
			}
		} else if ex.torch {
			next := ex
			next.time += 7
			next.torch = false
			s.add(next)
		}
		return ex

	}
	return explorer{}
}

type search struct {
	*scan
	been     []image.Point
	par      []int
	frontier []explorer
	seen     map[explorerState]int
	pend     map[explorerState]int
	best     explorer
	risk     map[image.Point]int
}

func (s *search) Len() int          { return len(s.frontier) }
func (s *search) Swap(i int, j int) { s.frontier[i], s.frontier[j] = s.frontier[j], s.frontier[i] }
func (s *search) Less(i int, j int) bool {
	return s.preferExplorer(s.frontier[i], s.frontier[j])
}

func (s *search) riskAt(p image.Point) int {
	const riskShadow = 3

	if r, def := s.risk[p]; def {
		return r
	}

	r := 0
	switch s.regionType(p) {
	case regionWet:
		r++
	case regionNarrow:
		r += 2
	}

	dt := signPoint(s.target.Sub(p))
	for np, i := p, riskShadow; i > 0 && np.Y != s.target.Y; i-- {
		np.Y += dt.Y
		np.X = p.X
		for i := riskShadow; i > 0 && np.X != s.target.X; i-- {
			np.X += dt.X
			r += s.riskAt(np)
		}
	}

	if s.risk == nil {
		s.risk = make(map[image.Point]int, 1024)
	}
	s.risk[p] = r

	return r
}

func signPoint(p image.Point) image.Point {
	if p.X < 0 {
		p.X = -1
	} else if p.X > 0 {
		p.X = 1
	}
	if p.Y < 0 {
		p.Y = -1
	} else if p.Y > 0 {
		p.Y = 1
	}
	return p
}

func (s *search) preferExplorer(a, b explorer) bool {
	da := manhattanDistance(a.Point, s.target)
	db := manhattanDistance(b.Point, s.target)

	da *= s.riskAt(a.Point)
	db *= s.riskAt(b.Point)

	// return da < db

	if da < db {
		return true
	}
	if db < da {
		return false
	}
	return a.time < b.time

	// if a.time < b.time {
	// 	return true
	// }
	// if a.time > b.time {
	// 	return false
	// }

	// td := s.target.X + s.target.Y
	// if da < 0 {
	// 	da *= da
	// }
	// if db < 0 {
	// 	db *= db
	// }
	// return da < db

}

func (s *search) add(next explorer) {
	if s.best.time > 0 && next.time > s.best.time {
		return
	}
	if t, def := s.seen[next.explorerState]; def && t < next.time {
		return
	}
	if t, def := s.pend[next.explorerState]; def && t < next.time {
		return
	}
	s.pend[next.explorerState] = next.time
	heap.Push(s, next)
}

func manhattanDistance(a, b image.Point) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (s *search) Push(x interface{}) {
	ex := x.(explorer)
	ex.parent = ex.id
	ex.id = len(s.been)
	s.been = append(s.been, ex.Point)
	s.par = append(s.par, ex.parent)
	s.frontier = append(s.frontier, ex)
}

func (s *search) Pop() interface{} {
	x := s.frontier[len(s.frontier)-1]
	s.frontier = s.frontier[:len(s.frontier)-1]
	return x
}

func (s *search) cull() (n int) {
	i := 0
	for j := 0; j < len(s.frontier); j++ {
		ex := s.frontier[j]

		if s.best.time > 0 && ex.time > s.best.time {
			n++
			continue
		}
		if t, def := s.seen[ex.explorerState]; def && t < ex.time {
			n++
			continue
		}

		if i != j {
			s.frontier[i] = s.frontier[j]
		}
		i++

	}
	if i < len(s.frontier) {
		s.frontier = s.frontier[:i]
		heap.Init(s)
	}
	return n
}

func run(in, out *os.File) error {
	var sc scan
	sc.depth = *depthFlag
	sc.target = image.Pt(*targetX, *targetY)

	min := image.ZP
	max := sc.target.Add(image.Pt(1, 1))

	// part 1
	// var buf bytes.Buffer
	risk := 0
	for p := min; p.Y < max.Y; p.Y++ {
		for p.X = min.X; p.X < max.X; p.X++ {
			rt := sc.regionType(p)
			switch rt {
			// case regionRocky:
			// 	buf.WriteByte('.')
			case regionWet:
				// buf.WriteByte('=')
				risk++
			case regionNarrow:
				// buf.WriteByte('|')
				risk += 2
			default:
				// buf.WriteByte('?')
			}
		}
		// buf.WriteByte('\n')
	}
	// buf.WriteTo(out)
	log.Printf("total risk: %v", risk)

	// part 2
	var s search
	s.scan = &sc
	s.been = make([]image.Point, 0, 1024*1024)
	s.par = make([]int, 0, 1024*1024)
	s.seen = make(map[explorerState]int, 1024*1024)
	s.pend = make(map[explorerState]int, 1024*1024)

	tick := time.NewTicker(time.Second)
	n := 0
	t0 := time.Now()

	heap.Push(&s, explorer{
		explorerState: explorerState{
			Point: image.ZP,
			torch: true,
			gear:  false,
		},
	})

	for {
		if s.Len() == 0 {
			break
		}

		ex := heap.Pop(&s).(explorer)
		delete(s.pend, ex.explorerState)

	deepen:
		if t, def := s.seen[ex.explorerState]; def && t < ex.time {
			continue
		}
		s.seen[ex.explorerState] = ex.time
		if s.best.time > 0 && ex.time > s.best.time {
			continue
		}

		select {
		case now := <-tick.C:
			log.Printf(
				"search % -40v (w/ % 8v rem, % 8v con, %.1f/s) best:%v",
				ex, s.Len(), n, float64(n)/(float64(now.Sub(t0))/float64(time.Second)),
				s.best.time,
			)
			if n := s.cull(); n > 0 {
				log.Printf("opp culled %v", n)
			}
		default:
		}

		n++
		if ex.done(&sc) {
			if s.best.time == 0 || ex.time < s.best.time {
				s.best = ex
				n := s.cull()
				log.Printf("FOUND %v culled:%v", ex, n)
			}
		} else if ex = ex.expand(&s); ex.time != 0 {
			if ex.time >= 200 {
				goto deepen
			}
			s.add(ex)
		}

	}

	return nil
}

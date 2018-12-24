package main

import (
	"aoc/internal/progprof"
	"container/heap"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"time"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
	defer progprof.Start()()
	anansi.MustRun(run(os.Stdin, os.Stdout))
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
	regionRocky regionType = iota
	regionWet
	regionNarrow
)

type equipType uint8

const (
	// In rocky regions, you can use the climbing gear or the torch.
	// You cannot use neither (you'll likely slip and fall).
	equipNothing equipType = iota // NOTE == regionRocky

	// In wet regions, you can use the climbing gear or neither tool.
	// You cannot use the torch (if it gets wet, you won't have a light source).
	equipTorch // NOTE == regionWet

	// In narrow regions, you can use the torch or neither tool.
	// You cannot use the climbing gear (it's too bulky to fit).
	equipGear // NOTE == regionNarrow
)

func (sc *scan) regionType(p image.Point) regionType {
	// If the erosion level modulo 3 is 0, the region's type is rocky.
	// If the erosion level modulo 3 is 1, the region's type is wet.
	// If the erosion level modulo 3 is 2, the region's type is narrow.
	return regionType(sc.erosionLevel(p) % 3)
}

var (
	depthFlag = flag.Int("d", 510, "depth")
	targetX   = flag.Int("x", 10, "target X")
	targetY   = flag.Int("y", 10, "target Y")
)

type explorerState struct {
	image.Point
	equip equipType
}

type explorer struct {
	explorerState
	time int
}

func (ex explorer) String() string {
	return fmt.Sprintf("t:%v @%v equip:%v", ex.time, ex.Point, ex.equip)
}

func (ex explorer) done(sc *scan) bool {
	return ex.Point == sc.target && ex.equip == equipTorch
}

func (ex explorer) expand(s *search) {
	// Finally, once you reach the target, you need the *torch* equipped before
	// you can find him in the dark. The target is always in a *rocky* region,
	// so if you arrive there with *climbing gear* equipped, you will need to
	// spend seven minutes switching to your torch.
	if ex.Point == s.target {
		if ex.equip != equipTorch {
			ex.equip = equipTorch
			ex.time += 7
			s.add(ex)
		}
		return
	}

	// You can change your currently equipped tool or put both away if your new
	// equipment would be valid for your current region. Switching to using the
	// climbing gear, torch, or neither always takes seven minutes, regardless
	// of which tools you start with. (For example, if you are in a rocky
	// region, you can switch from the torch to the climbing gear, but you
	// cannot switch to neither.)
	reg := s.regionType(ex.Point)
	for _, eq := range [...]equipType{equipNothing, equipTorch, equipGear} {
		if eq != ex.equip && eq != equipType(reg) {
			next := ex
			next.time += 7
			next.equip = eq
			s.add(next)
		}
	}

	// You can move to an adjacent region (up, down, left, or right; never
	// diagonally) if your currently equipped tool allows you to enter that
	// region. Moving to an adjacent region takes one minute. (For example,
	// if you have the torch equipped, you can move between rocky and
	// narrow regions, but cannot enter wet regions.)
	for _, dp := range [...]image.Point{
		image.Pt(0, -1), image.Pt(0, 1),
		image.Pt(1, 0), image.Pt(-1, 0),
	} {
		p := ex.Point.Add(dp)
		if p.X < 0 || p.Y < 0 {
			continue
		}
		if ex.equip == equipType(s.regionType(p)) {
			continue
		}

		next := ex
		next.time++
		next.Point = p
		s.add(next)
	}
}

type search struct {
	*scan
	frontier []explorer
	seen     map[explorerState]int
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
	if a.time < b.time {
		return true
	}
	if a.time > b.time {
		return false
	}
	if a.X < b.X {
		return true
	}
	if a.X > b.X {
		return false
	}
	if a.Y < b.Y {
		return true
	}
	if a.Y > b.Y {
		return false
	}
	return a.equip < b.equip
}

func (s *search) add(next explorer) {
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
	s.frontier = append(s.frontier, ex)
}

func (s *search) Pop() interface{} {
	x := s.frontier[len(s.frontier)-1]
	s.frontier = s.frontier[:len(s.frontier)-1]
	return x
}

func run(in, out *os.File) error {
	var sc scan
	sc.depth = *depthFlag
	sc.target = image.Pt(*targetX, *targetY)

	min := image.ZP
	max := sc.target.Add(image.Pt(1, 1))

	// part 1
	risk := 0
	for p := min; p.Y < max.Y; p.Y++ {
		for p.X = min.X; p.X < max.X; p.X++ {
			risk += int(sc.regionType(p))
		}
	}
	log.Printf("total risk: %v", risk)

	// part 2
	var s search
	s.scan = &sc
	s.seen = make(map[explorerState]int, 1024*1024)

	tick := time.NewTicker(time.Second)
	n := 0
	t0 := time.Now()

	heap.Push(&s, explorer{
		explorerState: explorerState{
			Point: image.ZP,
			equip: equipTorch,
		},
	})

	for s.Len() > 0 {
		ex := heap.Pop(&s).(explorer)

		if t, def := s.seen[ex.explorerState]; def && t <= ex.time {
			continue
		}
		s.seen[ex.explorerState] = ex.time
		if s.best.time > 0 && ex.time > s.best.time {
			continue
		}

		select {
		case now := <-tick.C:
			log.Printf(
				"search % -25v (w/ % 8v rem, % 8v con, %.1f/s) best:%v",
				ex, s.Len(), n, float64(n)/(float64(now.Sub(t0))/float64(time.Second)),
				s.best.time,
			)
		default:
		}

		n++
		if !ex.done(&sc) {
			ex.expand(&s)
		} else if s.best.time == 0 || ex.time < s.best.time {
			s.best = ex
			log.Printf("FOUND %v", ex)
		}
	}

	log.Printf(
		"searched % -25v (% 8v con, %.1f/s)",
		s.best, n, float64(n)/(float64(time.Now().Sub(t0))/float64(time.Second)),
	)

	return nil
}

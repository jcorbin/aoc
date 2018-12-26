package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"
	"unicode"

	"aoc/internal/display"
	"aoc/internal/geom"
	"aoc/internal/infernio"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	interactive = flag.Bool("i", false, "interactive mode")
	dump        = flag.Bool("d", false, "dump populated grid")
	region      = flag.Int("r", 0, "region threshold to sum (part 2)")
)

var errDone = errors.New("done")

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin))
}

var builtinInput = infernio.Builtin("" +
	"1, 1\n" +
	"1, 6\n" +
	"8, 3\n" +
	"3, 4\n" +
	"5, 5\n" +
	"8, 9\n")

func run(r io.Reader) (err error) {
	var prob ui

	if err := infernio.LoadInput(builtinInput, prob.load); err != nil {
		return err
	}

	prob.init()

	// interactive stepping of the search expansion
	if *interactive {
		return prob.interact()
	}

	// part 1
	if *region == 0 {
		if err := prob.populate(); err != nil {
			return err
		}

		if *dump {
			prob.render()
			if _, err := display.WriteGrid(os.Stdout, prob.g); err != nil {
				return err
			}
		}

		counts := prob.countArea()
		best, most := 0, 0
		for id, count := range counts {
			if best == 0 || count > most {
				best, most = id, count
			}
		}
		log.Printf(
			"the best is #%v(%q) @%v (off %v) with %v cells",
			best,
			prob.names[best],
			prob.points[best-1],
			prob.points[best-1].Sub(prob.Min),
			most,
		)
		return nil
	}

	// part 2
	n := 0
	limit := *region
	for pt := prob.Min; pt.Y < prob.Max.Y; pt.Y++ {
		var markID int
		if *dump {
			markID = len(prob.names)
			const v = 2 * colorOff / 3
			prob.names = append(prob.names, '#')
			prob.colors = append(prob.colors, ansi.RGB(v, v, v))
		}

		for pt.X = prob.Min.X; pt.X < prob.Max.X; pt.X++ {
			total := 0
			// var ds []int
			for _, loc := range prob.points {
				d := distance(pt, loc)
				// ds = append(ds, d)
				total += d
			}
			// log.Printf("@%v Σ %v = %v", pt, ds, total)
			if total < limit {
				if *dump {
					if i, _ := prob.Index(pt); prob.pointID[i] == 0 {
						prob.pointID[i] = markID
					}
				}
				n++
			}
		}
	}

	if *dump {
		prob.render()
		if _, err := display.WriteGrid(os.Stdout, prob.g); err != nil {
			return err
		}
	}

	log.Printf("%v cells within %v limit of all locs", n, limit)

	return nil
}

type problem struct {
	// input
	points []image.Point

	// world state
	geom.RCore
	interiors []int
	pointID   []int
	pointDist []int

	// processing
	frontier []cursor
}

type ui struct {
	problem

	haveInput anansi.InputSignal
	timer     *time.Timer
	expanding bool
	interval  time.Duration

	names  []rune
	colors []ansi.SGRColor
	g      anansi.Grid
}

type cursor struct {
	pt, origin image.Point
	i, id      int
}

func (cur cursor) String() string {
	return fmt.Sprintf(
		"@%v from %v i:%v id:%v",
		cur.pt, cur.origin,
		cur.i, cur.id,
	)
}

func (cur cursor) distance() (n int) {
	return distance(cur.origin, cur.pt)
}

func distance(a, b image.Point) (n int) {
	d := b.Sub(a)
	if d.X < 0 {
		d.X = -d.X
	}
	if d.Y < 0 {
		d.Y = -d.Y
	}
	return d.X + d.Y
}

func (prob *problem) init() {
	// compute bounding box
	prob.Rectangle = geom.PointRect(prob.points[0])
	for _, pt := range prob.points[1:] {
		prob.Rectangle = prob.Rectangle.Union(geom.PointRect(pt))
	}

	// collect interior point ids
	interior := prob.Rectangle.Inset(1)
	for i, pt := range prob.points {
		if pt.In(interior) {
			id := i + 1
			prob.interiors = append(prob.interiors, id)
		}
	}

	// setup state
	prob.Origin = prob.Min
	prob.Stride = prob.Dx()
	sz := prob.Size()
	prob.pointID = make([]int, sz.X*sz.Y)
	prob.pointDist = make([]int, sz.X*sz.Y)

	prob.frontier = prob.frontier[:0]
	prob.placePoints()
}

func (prob *problem) populate() (err error) {
	for len(prob.frontier) > 0 && err == nil {
		err = prob.expand()
	}
	if err == errDone {
		err = nil
	}
	return err
}

func (prob *problem) countArea() map[int]int {
	counts := make(map[int]int, len(prob.interiors))
	for _, id := range prob.interiors {
		counts[id] = 0
	}

	// prune interior points that escape to infinity
	n := len(prob.pointID)
	for i := 0; i < prob.Stride; i++ {
		delete(counts, prob.pointID[i])
	}
	for i := 0; i < n; i += prob.Stride {
		delete(counts, prob.pointID[i])
	}
	for i := prob.Stride - 1; i < n; i += prob.Stride {
		delete(counts, prob.pointID[i])
	}
	for i := n - prob.Stride; i < n; i++ {
		delete(counts, prob.pointID[i])
	}

	for _, id := range prob.pointID {
		if n, counted := counts[id]; counted {
			counts[id] = n + 1
		}
	}

	return counts
}

func (prob *problem) placePoints() {
	for i, pt := range prob.points {
		id := i + 1
		j, _ := prob.Index(pt)
		prob.pointID[j] = id
		prob.pointDist[j] = 0
		prob.frontier = append(prob.frontier, cursor{pt, pt, j, id})
	}
}

// expand the next cursor popped from the frontier in each cardinal direction.
func (prob *problem) expand() error {
	cur, ok := prob.pop()
	if !ok {
		return errDone
	}
	for _, move := range []image.Point{
		image.Pt(1, 0),
		image.Pt(0, 1),
		image.Pt(0, -1),
		image.Pt(-1, 0),
	} {
		if next, ok := prob.advance(cur, move); ok && prob.better(next) {
			prob.pointID[next.i] = next.id
			prob.pointDist[next.i] = next.distance()
			prob.frontier = append(prob.frontier, next)
		}
	}
	return nil
}

// pop returns the next, still valid, cursor from the frontier and true if one
// exists, the zero cursor and false otherwise.
func (prob *problem) pop() (cur cursor, _ bool) {
	for {
		if cur.id != 0 {
			if prob.valid(cur) {
				return cur, true
			}
		}
		i := len(prob.frontier) - 1
		if i < 0 {
			return cursor{}, false
		}
		cur = prob.frontier[i]
		prob.frontier = prob.frontier[:i]
	}
}

// valid returns true only if the given cursor should be further explored; if
// it returns false, then the cursor is pruned (skipped by pop).
func (prob *problem) valid(cur cursor) bool {
	return cur.id == prob.pointID[cur.i]
}

// better returns true only if the given cursor's id and distance are better
// than those already store at it's index.
func (prob *problem) better(cur cursor) bool {
	priorID := prob.pointID[cur.i]
	priorDist := prob.pointDist[cur.i]
	dist := cur.distance()
	if priorID > 0 {
		// break loop
		if priorID == cur.id {
			return false
		}
		// beaten by prior
		if dist > priorDist {
			return false
		}
		// nobody wins ties
		if dist == priorDist {
			prob.pointID[cur.i] = -1
			return false
		}
	} else if priorID < 0 {
		// prior tie still stands
		if dist >= priorDist {
			return false
		}
	}
	return true
}

// advance returns a copy of the given cursor moved in the given direction, and
// a boolean indicating whether that move was valid (true), or should instead
// be skipped (false).
func (prob *problem) advance(cur cursor, move image.Point) (_ cursor, ok bool) {
	cur.pt = cur.pt.Add(move)
	cur.i, ok = prob.Index(cur.pt)
	return cur, ok
}

func (prob *ui) interact() error {
	in, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	return anansi.NewTerm(in, os.Stdout,
		&prob.haveInput,
		anansi.RawMode,
	).RunWith(prob)
}

func (prob *ui) Run(term *anansi.Term) error {
	prob.timer = time.NewTimer(0)
	err := term.Loop(prob)
	if err == errDone {
		err = nil
	}
	return err
}

func (prob *ui) Update(term *anansi.Term) (redraw bool, _ error) {
	select {
	// user input
	case <-prob.haveInput.C:
		if _, err := term.ReadAny(); err != nil {
			return false, err
		}
		any := false
		for e, _, ok := term.Decode(); ok; e, _, ok = term.Decode() {
			switch e {
			// Ctrl-C to quit
			case 0x03:
				return false, errors.New("goodbye")

			// step with '.' key
			case ansi.Escape('.'):
				return true, prob.expand()

			// play/pause with <Space>
			case ansi.Escape(' '):
				if prob.expanding = !prob.expanding; !prob.expanding {
					if !prob.timer.Stop() {
						select {
						case <-prob.timer.C:
						default:
						}
					}
					fmt.Printf("paused.\r\n")
					return false, nil
				}
				fmt.Printf("playing...\r\n")
				any = true

			// speed control
			case ansi.Escape('-'):
				prob.interval *= 2
				any = true
			case ansi.Escape('+'):
				prob.interval /= 2
				any = true
			}
		}
		if !any {
			return false, nil
		}

	// timer tick, mostly when playing back; also used to draw initial frame.
	case <-prob.timer.C:

	}

	// advance the expansion, and (re)set the timer
	if prob.expanding {
		if err := prob.expand(); err != nil {
			return false, err
		}
		if prob.interval == 0 {
			prob.interval = time.Second / 10
		}
		prob.timer.Reset(prob.interval)
	}

	return true, nil
}

func (prob *ui) WriteTo(w io.Writer) (n int64, err error) {
	prob.render()
	return display.WriteGrid(w, prob.g)
}

func (prob *ui) init() {
	prob.problem.init()

	// assignNames
	prob.names = make([]rune, len(prob.points)+1)
	prob.colors = make([]ansi.SGRColor, len(prob.points)+1)
	prob.names[0] = '.'
	prob.colors[0] = ansi.RGB(colorOff/2, colorOff/2, colorOff/2)
	for i := range prob.points {
		id := i + 1
		prob.names[id] = rune(glyphs[i%len(glyphs)])
		prob.colors[id] = n2color(colorOff, i/len(glyphs))
	}
}

func (prob *ui) render() {
	prob.g.Resize(prob.Size())
	for pt := prob.Min; pt.Y < prob.Max.Y; pt.Y++ {
		pt.X = prob.Min.X
		i, _ := prob.Index(pt)
		j, _ := prob.g.CellOffset(ansi.PtFromImage(pt.Sub(prob.Origin)))
		for ; pt.X < prob.Max.X; pt.X++ {
			id := prob.pointID[i]
			dist := prob.pointDist[i]
			if id < 0 {
				id = 0
			}
			name := prob.names[id]
			attr := prob.colors[id].FG()
			if id != 0 {
				if dist == 0 {
					attr |= ansi.SGRAttrBold
				}
				if dist > 0 {
					name = unicode.ToLower(name)
					// attr |= n2color(32, dist).BG()
				}
			}
			prob.g.Rune[j] = name
			prob.g.Attr[j] = attr
			i++
			j++
		}
	}
}

var pointPattern = regexp.MustCompile(`^(\d+), *(\d+)$`)

func (prob *problem) load(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := pointPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			return fmt.Errorf("bad line %q, expecting %v", line, pointPattern)
		}
		var pt image.Point
		pt.X, _ = strconv.Atoi(parts[1])
		pt.Y, _ = strconv.Atoi(parts[2])
		prob.points = append(prob.points, pt)
	}
	return sc.Err()
}

const (
	glyphs     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	colorOff   = 128
	colorStep  = 8
	colorSteps = 16
)

func n2color(off, n int) ansi.SGRColor {
	var r, g, b uint8
	r = uint8(off + colorStep*(n%colorSteps))
	n /= colorSteps
	g = uint8(off + colorStep*(n%colorSteps))
	n /= colorSteps
	b = uint8(off + colorStep*(n%colorSteps))
	return ansi.RGB(r, g, b)
}

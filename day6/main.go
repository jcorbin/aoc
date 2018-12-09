package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	interactive = flag.Bool("i", false, "interactive mode")
	dump        = flag.Bool("d", false, "dump populated grid")
	region      = flag.Int("r", 0, "region threshold to sum (part 2)")
)

var errDone = errors.New("done")

var id2name func(id int) string

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin))
}

func run(r io.Reader) (err error) {
	var prob ui

	prob.points, err = readPoints(r)
	if err != nil {
		return err
	}

	prob.init()
	id2name = prob.id2name

	// interactive stepping of the search expansion
	if *interactive {
		return prob.interact()
	}

	// part 1
	if *region == 0 {
		if err := prob.populate(); err != nil {
			return err
		}
		log.Printf("populated in %v steps, skipped:%v, maxFrontierLen:%v", prob.step, prob.skip, prob.maxFL)

		if *dump {
			prob.render()
			if _, err := writeGrid(os.Stdout, prob.g); err != nil {
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
		if _, err := writeGrid(os.Stdout, prob.g); err != nil {
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
	RCore
	interiors []int
	pointID   []int
	pointDist []int

	// processing
	frontier queue
	step     int // counts cursors popped from the frontier
	skip     int // counts cursors skipped immediately after pop
	maxFL    int // max frontier length seen (measured after push)
}

type ui struct {
	problem

	haveInput anansi.InputSignal
	timer     *time.Timer
	expanding bool
	interval  time.Duration
	logBuf    bytes.Buffer

	names  []rune
	colors []ansi.SGRColor
	g      anansi.Grid
}

type cursor struct {
	pt, origin image.Point
	i, id      int
}

func (cur cursor) String() string {
	if id2name == nil {
		return fmt.Sprintf(
			"#%v @%v from %v i:%v",
			cur.id,
			cur.pt, cur.origin,
			cur.i,
		)
	}
	return fmt.Sprintf(
		"%s @%v from %v i:%v",
		id2name(cur.id),
		cur.pt, cur.origin,
		cur.i,
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
	prob.Min = prob.points[0]
	prob.Max = prob.points[1]
	for _, pt := range prob.points[1:] {
		if prob.Min.X > pt.X {
			prob.Min.X = pt.X
		}
		if prob.Min.Y > pt.Y {
			prob.Min.Y = pt.Y
		}
		if prob.Max.X < pt.X {
			prob.Max.X = pt.X
		}
		if prob.Max.Y < pt.Y {
			prob.Max.Y = pt.Y
		}
	}

	for i, pt := range prob.points {
		if pt.X == prob.Min.X {
			continue
		}
		if pt.Y == prob.Min.Y {
			continue
		}
		if pt.X == prob.Max.X {
			continue
		}
		if pt.Y == prob.Max.Y {
			continue
		}
		id := i + 1
		prob.interiors = append(prob.interiors, id)
	}

	prob.Max.X++
	prob.Max.Y++

	// setup state
	prob.Origin = prob.Min
	prob.Stride = prob.Dx()
	sz := prob.Size()
	prob.pointID = make([]int, sz.X*sz.Y)
	prob.pointDist = make([]int, sz.X*sz.Y)

	prob.frontier.Clear()
	prob.placePoints()
}

func (prob *problem) populate() (err error) {
	for prob.frontier.Len() > 0 && err == nil {
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
		prob.frontier.push(cursor{pt, pt, j, id})
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
			heap.Push(&prob.frontier, next)
			if n := prob.frontier.Len(); prob.maxFL < n {
				prob.maxFL = n
			}
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
			prob.skip++
		}
		if prob.frontier.Len() == 0 {
			return cursor{}, false
		}
		cur = heap.Pop(&prob.frontier).(cursor)
		prob.step++
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

type queue struct {
	cs []cursor
}

func (q *queue) Clear()   { q.cs = q.cs[:0] }
func (q *queue) Len() int { return len(q.cs) }

func (q *queue) push(cur cursor) { q.cs = append(q.cs, cur) }
func (q *queue) pop() cursor {
	i := len(q.cs) - 1
	cur := q.cs[i]
	q.cs = q.cs[:i]
	return cur
}

func (q *queue) Swap(i int, j int) { q.cs[i], q.cs[j] = q.cs[j], q.cs[i] }
func (q *queue) Less(i int, j int) bool {
	// TODO better to have a cursor.gen int field, so that we correctly
	// de-prioritizing back-tracking (which could also be pruned of course)
	di, dj := q.cs[i].distance(), q.cs[j].distance()
	if di == dj {
		return q.cs[i].id < q.cs[j].id
	}
	return di < dj
}
func (q *queue) Push(x interface{}) { q.push(x.(cursor)) }
func (q *queue) Pop() interface{}   { return q.pop() }

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

func (prob *ui) expand() error {
	err := prob.problem.expand()
	log.Printf("step:%v, skipped:%v, maxFrontierLen:%v", prob.step, prob.skip, prob.maxFL)
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

func (prob *ui) WriteTo(w io.Writer) (int64, error) {
	prob.render()
	logBytes := prob.logBuf.Bytes()
	li := bytes.IndexByte(logBytes, '\n')

	var buf anansi.Buffer

	writeLogLine := func() {
		if len(logBytes) == 0 {
			return
		}
		buf.WriteString(" ")
		if li < 0 {
			buf.Write(logBytes)
			logBytes = nil
			return
		}
		buf.Write(logBytes[:li])

		logBytes = logBytes[li+1:]
		li = bytes.IndexByte(logBytes, '\n')
	}

	buf.WriteString(strings.Repeat("-", prob.g.Rect.Dx()))
	writeLogLine()
	buf.WriteString("\r\n")

	var cur anansi.CursorState
	cur.Point = ansi.Pt(1, 1)
	for y := prob.g.Rect.Min.Y; y < prob.g.Rect.Max.Y; y++ {
		cur = writeGridRow(&buf, cur, prob.g, y)
		writeLogLine()
		buf.WriteString("\r\n")
		cur.X = 1
		cur.Y++
	}
	_, _ = buf.WriteTo(w)

	for len(logBytes) > 0 {
		buf.WriteString(strings.Repeat(" ", prob.g.Rect.Dx()))
		writeLogLine()
		buf.WriteString("\r\n")
	}

	n, err := buf.WriteTo(w)
	if err != nil {
		return n, err
	}
	prob.logBuf.Reset()
	return n, nil
}

func (prob *ui) init() {
	prob.problem.init()

	log.SetOutput(&prob.logBuf)
	log.SetFlags(0 /* log.Ltime | log.Lmicroseconds */)

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

func (prob *ui) id2name(id int) string {
	if id >= 0 && id < len(prob.names) {
		return string(prob.names[id])
	}
	return fmt.Sprintf("#%d", id)
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

func readPoints(r io.Reader) (points []image.Point, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := pointPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			log.Printf("NO MATCH %q", line)
			continue
		}

		var pt image.Point
		pt.X, _ = strconv.Atoi(parts[1])
		pt.Y, _ = strconv.Atoi(parts[2])
		points = append(points, pt)
	}
	return points, sc.Err()
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

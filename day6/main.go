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

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	interactive = flag.Bool("i", false, "interactive mode")
	dump        = flag.Bool("d", false, "dump populated grid")
)

func main() {
	flag.Parse()
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

func run(r io.Reader) error {
	points, err := readPoints(r)
	if err != nil {
		return err
	}

	var prob ui

	if *interactive {
		prob.points = points
		prob.init()
		return prob.interact()
	}

	prob.points = points
	prob.init()
	if err := prob.populate(); err != nil {
		return err
	}

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

type problem struct {
	// input
	points []image.Point

	// world state
	RCore
	interiors []int
	pointID   []int
	pointDist []int

	// processing
	frontier []cursor
}

type ui struct {
	problem

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
	d := cur.pt.Sub(cur.origin)
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

func (prob *problem) pop() cursor {
	cur := prob.frontier[len(prob.frontier)-1]
	prob.frontier = prob.frontier[:len(prob.frontier)-1]
	return cur
}

var errDone = errors.New("done")

func (prob *problem) expand() error {
	if len(prob.frontier) == 0 {
		return errDone
	}
	cur := prob.pop()
	// skip already processed cursors
	for prob.pointID[cur.i] != cur.id {
		// log.Printf("skip %v", cur)
		if len(prob.frontier) == 0 {
			return errDone
		}
		cur = prob.pop()
	}
	// log.Printf("#%v expanding %v priorID:%v priorDist:%v",
	// 	len(prob.frontier), cur, prob.pointID[cur.i], prob.pointDist[cur.i])

	// expand in cardinal directions
	for _, dir := range []image.Point{
		image.Pt(1, 0),
		image.Pt(0, 1),
		image.Pt(0, -1),
		image.Pt(-1, 0),
	} {
		var in bool
		next := cur
		next.pt = next.pt.Add(dir)
		if next.i, in = prob.Index(next.pt); in {
			priorID := prob.pointID[next.i]
			priorDist := prob.pointDist[next.i]
			dist := next.distance()

			if priorID > 0 {

				// break loop
				if priorID == next.id {
					continue
				}

				// beaten by prior
				if dist > priorDist {
					continue
				}

				// nobody wins ties
				if dist == priorDist {
					// log.Printf("%v tied with #%v(%v)", next, priorID, priorDist)
					prob.pointID[next.i] = -1
					continue
				}

			} else if priorID < 0 {
				// prior tie still stands
				if dist >= priorDist {
					continue
				}
			}

			// log.Printf(
			// 	"@%v replace #%v(%v) with #%v(%v)",
			// 	next.pt,
			// 	priorID, priorDist,
			// 	next.id, dist,
			// )

			prob.pointID[next.i] = next.id
			prob.pointDist[next.i] = dist
			prob.frontier = append(prob.frontier, next)
		}
	}

	return nil
}

func (prob *ui) interact() error {
	in, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}

	haveInput := anansi.InputSignal{}
	timer := time.NewTimer(0)

	term := anansi.NewTerm(in, os.Stdout, &haveInput)
	term.SetRaw(true)
	term.SetEcho(false)

	return term.RunWith(func(term *anansi.Term) error {
		expanding := false
		interval := time.Second / 10

		for {
			select {
			case <-haveInput.C:
				if _, err := term.ReadAny(); err != nil {
					return err
				}
				for e, _, ok := term.Decode(); ok; e, _, ok = term.Decode() {
					switch e {
					case 0x03:
						return errors.New("goodbye")

					case ansi.Escape('.'):
						if err := prob.expand(); err != nil {
							return err
						}

					case ansi.Escape(' '):
						expanding = !expanding
						if expanding {
							fmt.Printf("playing...\r\n")
						} else {
							timer.Stop()
							fmt.Printf("paused.\r\n")
						}

					case ansi.Escape('-'):
						interval *= 2
					case ansi.Escape('+'):
						interval /= 2

					}
				}

			case <-timer.C:
				if expanding {
					if err := prob.expand(); err != nil {
						return err
					}
				}
			}

			prob.render()

			if _, err := writeGrid(os.Stdout, prob.g); err != nil {
				return err
			}
			if expanding {
				timer.Reset(interval)
			}
		}

	})
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

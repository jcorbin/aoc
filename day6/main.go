package main

import (
	"bufio"
	"errors"
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

func main() {
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

var (
	// world state
	rc        RCore
	pointID   []int
	pointDist []int

	// processing
	frontier []cursor

	// output
	names  []rune
	colors []ansi.SGRColor
	g      anansi.Grid
)

type cursor struct {
	image.Point
	i, id, dist int
}

func (cur cursor) String() string {
	return fmt.Sprintf("@%v i:%v id:%v dist:%v", cur.Point, cur.i, cur.id, cur.dist)
}

func run(r io.Reader) error {
	// read points
	points, err := readPoints(r)
	if err != nil {
		return err
	}

	setup(points)
	placePoints(points)
	assignNames(points)

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
						if err := expand(); err != nil {
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
					if err := expand(); err != nil {
						return err
					}
				}
			}

			if err := render(); err != nil {
				return err
			}
			if expanding {
				timer.Reset(interval)
			}
		}

	})
}

func setup(points []image.Point) {
	// compute bounding box
	rc.Min = points[0]
	rc.Max = points[1]
	for _, pt := range points[1:] {
		if rc.Min.X > pt.X {
			rc.Min.X = pt.X
		}
		if rc.Min.Y > pt.Y {
			rc.Min.Y = pt.Y
		}
		if rc.Max.X < pt.X {
			rc.Max.X = pt.X
		}
		if rc.Max.Y < pt.Y {
			rc.Max.Y = pt.Y
		}
	}
	rc.Max.X++
	rc.Max.Y++

	// setup state
	rc.Origin = rc.Min
	rc.Stride = rc.Dx()
	sz := rc.Size()
	pointID = make([]int, sz.X*sz.Y)
	pointDist = make([]int, sz.X*sz.Y)
}

func placePoints(points []image.Point) {
	for i, pt := range points {
		id := i + 1
		j, _ := rc.Index(pt)
		pointID[j] = id
		pointDist[j] = 0
		frontier = append(frontier, cursor{pt, j, id, 0})
	}
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

func assignNames(points []image.Point) {
	names = make([]rune, len(points)+1)
	colors = make([]ansi.SGRColor, len(points)+1)
	names[0] = '.'
	colors[0] = ansi.RGB(colorOff/2, colorOff/2, colorOff/2)
	for i := range points {
		id := i + 1
		names[id] = rune(glyphs[i%len(glyphs)])
		colors[id] = n2color(colorOff, i/len(glyphs))
	}
}

func pop() cursor {
	cur := frontier[len(frontier)-1]
	frontier = frontier[:len(frontier)-1]
	return cur
}

func expand() error {
	if len(frontier) == 0 {
		return errors.New("done")
	}
	cur := pop()
	// skip already processed cursors
	for pointID[cur.i] != cur.id {
		if len(frontier) == 0 {
			return errors.New("done")
		}
		cur = pop()
		fmt.Printf("skip %v\r\n", cur)
	}
	fmt.Printf("#%v expanding %v priorID:%v priorDist:%v\r\n",
		len(frontier), cur, pointID[cur.i], pointDist[cur.i])

	// expand in cardinal directions
	for _, dir := range []image.Point{
		image.Pt(1, 0),
		image.Pt(0, 1),
		image.Pt(0, -1),
		image.Pt(-1, 0),
	} {
		var in bool
		next := cur
		next.Point = next.Point.Add(dir)
		next.dist++
		if next.i, in = rc.Index(next.Point); in {
			priorID := pointID[next.i]
			priorDist := pointDist[next.i]

			if priorID > 0 {

				// break loop
				if priorID == next.id {
					continue
				}

				// beaten by prior
				if next.dist > priorDist {
					continue
				}

				// nobody wins ties
				if next.dist == priorDist {
					pointID[next.i] = -1
					continue
				}

			} else if priorID < 0 {
				// XXX can we beat a prior tie?
				continue
			}

			pointID[next.i] = next.id
			pointDist[next.i] = next.dist
			frontier = append(frontier, next)
		}
	}

	return nil
}

func render() error {
	g.Resize(rc.Size())
	for pt := rc.Min; pt.Y < rc.Max.Y; pt.Y++ {
		pt.X = rc.Min.X
		i, _ := rc.Index(pt)
		j, _ := g.CellOffset(ansi.PtFromImage(pt.Sub(rc.Origin)))
		for ; pt.X < rc.Max.X; pt.X++ {
			id := pointID[i]
			dist := pointDist[i]
			if id < 0 {
				id = 0
			}
			name := names[id]
			c := colors[id].FG()
			if id != 0 && dist > 0 {
				name = unicode.ToLower(name)
				// c |= n2color(32, dist).BG()
			}
			g.Rune[j] = name
			g.Attr[j] = c
			i++
			j++
		}
	}

	_, err := writeGrid(os.Stdout, g)
	return err
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

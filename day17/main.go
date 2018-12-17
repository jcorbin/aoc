package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/aoc/internal/geom"
)

var (
	verbose = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type spec struct {
	x [2]int
	y [2]int
}

type world struct {
	geom.RCore
	d []byte
}

func run(in, out *os.File) error {
	specs, err := read(in)
	if err != nil {
		return err
	}

	// part 1
	var w world

	springAt := image.Pt(500, 0)

	var bounds image.Rectangle

	// compute bounds
	if len(specs) > 0 {
		s := specs[0]
		bounds = image.Rect(s.x[0], s.y[0], s.x[1]+1, s.y[1]+1)
	}
	for _, s := range specs[1:] {
		specRect := image.Rect(s.x[0], s.y[0], s.x[1]+1, s.y[1]+1)
		bounds = bounds.Union(specRect)
	}

	// padding for anything butted up against the edge
	bounds.Min.X--
	bounds.Max.X++

	// allocate
	sr := image.Rectangle{Min: springAt, Max: springAt.Add(image.Pt(1, 1))}
	w.Rectangle = bounds.Union(sr)
	w.Stride = w.Rectangle.Dx()
	w.Origin = w.Rectangle.Min
	w.d = make([]byte, w.Stride*w.Rectangle.Dy())

	// fill with sand
	for i := range w.d {
		w.d[i] = '.'
	}

	// place clay
	for _, s := range specs {
		if s.x[0] == s.x[1] { // vertical
			y := s.y[0]
			i, _ := w.Index(image.Pt(s.x[0], y))
			for ; y <= s.y[1]; y++ {
				w.d[i] = '#'
				i += w.Stride
			}
		} else if s.y[0] == s.y[1] { // horizontal
			x := s.x[0]
			i, _ := w.Index(image.Pt(x, s.y[0]))
			for ; x <= s.x[1]; x++ {
				w.d[i] = '#'
				i++
			}
		} else {
			panic("inconceivable: malformed spec data")
		}
	}

	if i, ok := w.Index(springAt); ok {
		w.d[i] = '+'
	} else {
		panic("inconceivable: spring not in range")
	}

	dumpI := 0
	dump := func() error {
		dumpI++
		var buf bytes.Buffer
		buf.Grow(1024)
		fmt.Fprintf(&buf, "DUMP %v %v\n", dumpI, w.Rectangle)

		nXDigits := 0
		for x := w.Rectangle.Max.X; x > 0; x /= 10 {
			nXDigits++
		}

		nYDigits := 0
		for y := w.Rectangle.Max.Y; y > 0; y /= 10 {
			nYDigits++
		}

		for i := 0; i < nXDigits; i++ {
			fmt.Fprintf(&buf, "% *s ", nYDigits+1, "")
			for x := w.Rectangle.Min.X; x < w.Rectangle.Max.X; x++ {
				xd := x
				for j := nXDigits - 1; j > i; j-- {
					xd /= 10
				}
				buf.WriteByte('0' + byte(xd%10))
			}
			buf.WriteByte('\n')
		}

		for pt := w.Rectangle.Min; pt.Y < w.Rectangle.Max.Y; pt.Y++ {
			fmt.Fprintf(&buf, "% *d ", nYDigits+1, pt.Y)
			for pt.X = w.Rectangle.Min.X; pt.X < w.Rectangle.Max.X; pt.X++ {
				i, _ := w.Index(pt)
				buf.WriteByte(w.d[i])
			}
			buf.WriteByte('\n')
		}

		for i := 0; i < nXDigits; i++ {
			fmt.Fprintf(&buf, "% *s ", nYDigits+1, "")
			for x := w.Rectangle.Min.X; x < w.Rectangle.Max.X; x++ {
				xd := x
				for j := nXDigits - 1; j > i; j-- {
					xd /= 10
				}
				buf.WriteByte('0' + byte(xd%10))
			}
			buf.WriteByte('\n')
		}

		_, err := buf.WriteTo(out)
		return err
	}

	fillWith := func(from, to image.Point, d byte) (any bool) {
		for fill := from.Add(image.Pt(1, 0)); fill.X < to.X; fill.X++ {
			i, _ := w.Index(fill)
			if w.d[i] != d {
				w.d[i] = d
				any = true
			}
		}
		return any
	}

	scanLeft := func(pt image.Point) (image.Point, bool) {
		j, _ := w.Index(pt)
		for x := pt.X; x >= w.Rectangle.Min.X; {
			if dn := w.d[j+w.Stride]; dn == '.' || dn == '|' {
				return image.Pt(x, pt.Y), true
			}
			if w.d[j] == '#' {
				return image.Pt(x, pt.Y), false
			}
			j--
			x--
		}
		panic("inconceivable: should either escape or block")
	}

	scanRight := func(pt image.Point) (image.Point, bool) {
		j, _ := w.Index(pt)
		for x := pt.X; x < w.Rectangle.Max.X; {
			if dn := w.d[j+w.Stride]; dn == '.' || dn == '|' {
				return image.Pt(x, pt.Y), true
			}
			if w.d[j] == '#' {
				return image.Pt(x, pt.Y), false
			}
			j++
			x++
		}
		panic("inconceivable: should either escape or block")
	}

	frontier := make([]image.Point, 0, 1024)
	frontier = append(frontier, springAt.Add(image.Pt(0, 1)))

	// dump()

	for sanity := 100000; len(frontier) > 0; sanity-- {
		if sanity < 0 {
			dump()
			return fmt.Errorf("sanity exhausted %v", frontier)
		}

		if *verbose {
			if err := dump(); err != nil {
				return err
			}
		}

		start := frontier[len(frontier)-1]
		frontier = frontier[:len(frontier)-1]

		if *verbose {
			log.Printf("start @%v", start)
		}
		pt := start
		for placed := false; !placed; {
			i, ok := w.Index(pt)
			if !ok {
				break
			}

			iDown := i + w.Stride
			if iDown >= len(w.d) {
				w.d[i] = '|'
				break
			}

			switch w.d[iDown] {
			case '.':
				fallthrough
			case '|':
				w.d[i] = '|'
				pt.Y++

			case '~', '#':
				left, leftEscape := scanLeft(pt)
				right, rightEscape := scanRight(pt)

				if leftEscape || rightEscape {
					if leftEscape {
						if *verbose {
							log.Printf("escape left @%v", left)
						}
						l, r := left.Add(image.Pt(-1, 0)), pt.Add(image.Pt(1, 0))
						if !rightEscape {
							r = right
						}
						if fillWith(l, r, '|') {
							frontier = append(frontier, left)
						}
					}
					if rightEscape {
						if *verbose {
							log.Printf("escape right @%v", right)
						}
						l, r := pt, right.Add(image.Pt(1, 0))
						if !leftEscape {
							l = left
						}
						if fillWith(l, r, '|') {
							frontier = append(frontier, right)
						}
					}
				} else {
					if *verbose {
						log.Printf("filling @%v", pt)
					}
					if fillWith(left, right, '~') {
						frontier = append(frontier, start)
					} else if pt == start {
						start.Y--
						pt.Y--
						continue
					}
				}
				placed = true
			}
		}
	}

	if err := dump(); err != nil {
		return err
	}

	n := 0
	for pt := bounds.Min; pt.Y < bounds.Max.Y; pt.Y++ {
		for pt.X = bounds.Min.X; pt.X < bounds.Max.X; pt.X++ {
			i, _ := w.Index(pt)
			switch w.d[i] {
			case '|', '~':
				n++
			}
		}
	}

	log.Printf("HAVE %v", n)

	// part 2
	// TODO

	return nil
}

var specPattern = regexp.MustCompile("" +
	`(?:^x=(\d+), *y=(\d+)\.\.(\d+)$)` +
	`|` +
	`(?:^y=(\d+), *x=(\d+)\.\.(\d+)$)`,
)

func read(r io.Reader) (specs []spec, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := specPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			return specs, fmt.Errorf("unrecognized line %q", line)
		}

		var s spec

		if parts[1] != "" {
			s.x[0], _ = strconv.Atoi(parts[1])
			s.y[0], _ = strconv.Atoi(parts[2])
			s.y[1], _ = strconv.Atoi(parts[3])
			s.x[1] = s.x[0]
		} else if parts[4] != "" {
			s.y[0], _ = strconv.Atoi(parts[4])
			s.x[0], _ = strconv.Atoi(parts[5])
			s.x[1], _ = strconv.Atoi(parts[6])
			s.y[1] = s.y[0]
		} else {
			panic("inconceivable: exhaustive pattern")
		}

		specs = append(specs, s)
	}
	return specs, sc.Err()
}

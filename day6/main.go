package main

import (
	"bufio"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"unicode"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

func main() {
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

func run(r io.Reader) error {
	// read points
	points, err := readPoints(r)
	if err != nil {
		return err
	}

	// compute bounding box
	var rc RCore
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
	pointID := make([]int, sz.X*sz.Y)
	pointDist := make([]int, sz.X*sz.Y)

	// place points
	for i, pt := range points {
		id := i + 1
		j, _ := rc.Index(pt)
		pointID[j] = id
	}

	// assign marks
	const glyphs = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	names := make([]rune, len(points)+1)
	colors := make([]ansi.SGRColor, len(points)+1)
	names[0] = '.'
	colors[0] = ansi.RGB(64, 64, 64)
	for i := range points {
		id := i + 1
		names[id] = rune(glyphs[i%len(glyphs)])

		const (
			colorOff   = 128
			colorStep  = 8
			colorSteps = 16
		)
		n := i / len(glyphs)

		var r, g, b uint8
		r = uint8(colorOff + colorStep*(n%colorSteps))
		n /= colorSteps
		g = uint8(colorOff + colorStep*(n%colorSteps))
		n /= colorSteps
		b = uint8(colorOff + colorStep*(n%colorSteps))
		colors[id] = ansi.RGB(r, g, b)
		// TODO append to frontier
	}

	// TODO expand frontier

	// output rendering
	var g anansi.Grid
	g.Resize(sz)
	for pt := rc.Min; pt.Y < rc.Max.Y; pt.Y++ {
		pt.X = rc.Min.X
		i, _ := rc.Index(pt)
		j, _ := g.CellOffset(ansi.PtFromImage(pt.Sub(rc.Origin)))
		for ; pt.X < rc.Max.X; pt.X++ {
			id := pointID[i]
			name := names[id]
			if id != 0 && pointDist[i] > 0 {
				name = unicode.ToLower(name)
			}
			g.Rune[j] = name
			g.Attr[j] = colors[id].FG()
			i++
			j++
		}
	}

	_, err = writeGrid(os.Stdout, g)
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

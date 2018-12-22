package main

import (
	"bytes"
	"flag"
	"image"
	"log"
	"os"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
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
	log.Printf("WUT @%v %v", p, el)
	return regionInvalid
}

var (
	depthFlag = flag.Int("d", 510, "depth")
	targetX   = flag.Int("x", 10, "target X")
	targetY   = flag.Int("y", 10, "target Y")
)

func run(in, out *os.File) error {
	var sc scan
	sc.depth = *depthFlag
	sc.target = image.Pt(*targetX, *targetY)

	min := image.ZP
	max := sc.target.Add(image.Pt(1, 1))

	// part 1
	var buf bytes.Buffer

	risk := 0
	for p := min; p.Y < max.Y; p.Y++ {
		for p.X = min.X; p.X < max.X; p.X++ {
			rt := sc.regionType(p)
			switch rt {
			case regionRocky:
				buf.WriteByte('.')
			case regionWet:
				buf.WriteByte('=')
				risk++
			case regionNarrow:
				buf.WriteByte('|')
				risk += 2
			default:
				buf.WriteByte('?')
			}
		}
		buf.WriteByte('\n')
	}
	buf.WriteTo(out)
	log.Printf("total risk: %v", risk)

	// part 2
	// TODO

	return nil
}

package main

import (
	"flag"
	"image"
	"log"
	"os"

	"github.com/jcorbin/anansi"

	"github.com/jcorbin/aoc/internal/geom"
)

var (
	serialFlag = flag.Int("serial", 0, "serial sumber")
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

// The power level in a given fuel cell can be found through the following
// process:
func powerLevel(pt image.Point, serial int) int {
	// Find the fuel cell\'s *rack ID*, which is its *X coordinate plus 10*.
	rackID := pt.X + 10

	// Begin with a power level of the *rack ID* times the *Y coordinate*.
	lvl := rackID * pt.Y

	// Increase the power level by the value of the *grid serial number* (your puzzle input).
	lvl += serial

	// Set the power level to itself multiplied by the *rack ID*.
	lvl *= rackID

	// Keep only the *hundreds digit* of the power level (so `12345` becomes `3`; numbers with no hundreds digit become `0`).
	lvl /= 100
	lvl %= 10

	// *Subtract 5* from the power level.
	lvl -= 5

	return lvl
}

type fuelGrid struct {
	geom.RCore
	d []int
}

func run(in, out *os.File) error {
	serial := *serialFlag

	// build power grid
	var fg fuelGrid
	fg.Min = image.Pt(1, 1)
	fg.Max = image.Pt(301, 301)
	fg.Stride = 300
	fg.Origin = fg.Min
	fg.d = make([]int, fg.Dx()*fg.Dy())
	for pt := fg.Min; pt.Y < fg.Max.Y; pt.Y++ {
		for pt.X = fg.Min.X; pt.X < fg.Max.X; pt.X++ {
			i, _ := fg.Index(pt)
			fg.d[i] = powerLevel(pt, serial)
		}
	}

	sq := 3

	// part 1
	cfg := accumulate(fg, sq)
	best, level := maxCell(cfg)
	best = best.Add(image.Pt(1-sq, 1-sq)) // FIXME due to stencil structure
	log.Printf("found %v @%v sz:3", level, best)

	// part 2
	var bestSize int
	best, level = image.ZP, 0
	for sq := fg.Min.X; sq < fg.Max.X; sq++ {
		cfg := accumulate(fg, sq)
		subBest, subLevel := maxCell(cfg)
		subBest = subBest.Add(image.Pt(1-sq, 1-sq)) // FIXME due to stencil structure
		log.Printf("sq:%v subBest:%v subLevel:%v", sq, subBest, subLevel)
		if level < subLevel {
			best, level = subBest, subLevel
			bestSize = sq
			log.Printf("have best:%v level:%v size:%v", best, level, bestSize)
		}
	}
	log.Printf("found %v @%v sz:%v", level, best, bestSize)

	return nil
}

func maxCell(fg fuelGrid) (best image.Point, level int) {
	for pt := fg.Min; pt.Y < fg.Max.Y-2; pt.Y++ {
		for pt.X = fg.Min.X; pt.X < fg.Max.X-2; pt.X++ {
			i, _ := fg.Index(pt)
			if v := fg.d[i]; level < v {
				best, level = pt, v
			}
		}
	}
	return best, level
}

func accumulate(fg fuelGrid, sq int) fuelGrid {
	// spread stencil
	stencil := make([]image.Point, 0, sq*sq)
	for i := 0; i < sq; i++ {
		for j := 0; j < sq; j++ {
			stencil = append(stencil, image.Pt(i, j))
		}
	}

	var cfg fuelGrid
	cfg.Stride = fg.Stride
	cfg.Rectangle = fg.Rectangle
	cfg.Stride = 300
	cfg.Origin = fg.Min
	cfg.d = make([]int, cfg.Dx()*cfg.Dy())
	for pt := fg.Min; pt.Y < fg.Max.Y; pt.Y++ {
		for pt.X = fg.Min.X; pt.X < fg.Max.X; pt.X++ {
			i, _ := fg.Index(pt)
			v := fg.d[i]
			for _, dpt := range stencil {
				if j, ok := cfg.Index(pt.Add(dpt)); ok {
					cfg.d[j] += v
				}
			}
		}
	}
	return cfg
}

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

type solver interface {
	solve(size int) (loc image.Point, level int)
}

func fuelGridSolver(serial int, bounds image.Rectangle) solver {
	return buildFuelGrid(serial, bounds)
}

var factory = fuelGridSolver

func run(in, out *os.File) error {
	bounds := image.Rect(1, 1, 301, 301)
	solver := factory(*serialFlag, bounds)

	// part 1
	best, level := solver.solve(3)
	log.Printf("found %v @%v sz:3", level, best)

	// part 2
	var bestSize int
	best, level = image.ZP, 0
	for size := bounds.Min.X; size < bounds.Max.X; size++ {
		subBest, subLevel := solver.solve(size)
		log.Printf("size:%v subBest:%v subLevel:%v", size, subBest, subLevel)
		if level < subLevel {
			best, level = subBest, subLevel
			bestSize = size
			log.Printf("have best:%v level:%v size:%v", best, level, bestSize)
		}
	}
	log.Printf("found %v @%v sz:%v", level, best, bestSize)

	return nil
}

type fuelGrid struct {
	geom.RCore
	d []int
}

func buildFuelGrid(serial int, bounds image.Rectangle) (fg fuelGrid) {
	fg.Rectangle = bounds
	fg.Stride = fg.Dx()
	fg.Origin = fg.Min
	fg.d = make([]int, fg.Stride*fg.Dy())
	for pt := fg.Min; pt.Y < fg.Max.Y; pt.Y++ {
		for pt.X = fg.Min.X; pt.X < fg.Max.X; pt.X++ {
			i, _ := fg.Index(pt)
			fg.d[i] = powerLevel(pt, serial)
		}
	}
	return fg
}

func (fg fuelGrid) solve(size int) (loc image.Point, level int) {
	cfg := accumulate(fg, size)
	loc, level = maxCell(cfg)
	loc = loc.Add(image.Pt(1-size, 1-size)) // FIXME due to stencil structure
	return loc, level
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

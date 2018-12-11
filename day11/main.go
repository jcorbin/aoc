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
	verbose = flag.Bool("v", false, "enable verbose output")
	serial  = flag.Int("serial", 0, "serial sumber")
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

func specSolver(serial int, bounds image.Rectangle) solver {
	return spec{serial, bounds}
}

func partialSumSolver(serial int, bounds image.Rectangle) solver {
	return buildPartialSums(serial, bounds)
}

var factory = partialSumSolver

func run(in, out *os.File) error {
	bounds := image.Rect(1, 1, 301, 301)
	solver := factory(*serial, bounds)

	// part 1
	best, level := solver.solve(3)
	log.Printf("found %v @%v sz:3", level, best)

	// part 2
	var bestSize int
	best, level = image.ZP, 0
	for size := bounds.Min.X; size < bounds.Max.X; size++ {
		subBest, subLevel := solver.solve(size)
		if *verbose {
			log.Printf("size:%v subBest:%v subLevel:%v", size, subBest, subLevel)
		}
		if level < subLevel {
			best, level = subBest, subLevel
			bestSize = size
			if *verbose {
				log.Printf("have best:%v level:%v size:%v", best, level, bestSize)
			}
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
	for pt := fg.Min; pt.Y < fg.Max.Y-size; pt.Y++ {
		for pt.X = fg.Min.X; pt.X < fg.Max.X-size; pt.X++ {
			total := 0
			for dy := 0; dy < size; dy++ {
				for dx := 0; dx < size; dx++ {
					i, _ := fg.Index(pt.Add(image.Pt(dx, dy)))
					total += fg.d[i]
				}
			}
			if level < total {
				loc, level = pt, total
			}
		}
	}
	return loc, level
}

type spec struct {
	serial int
	image.Rectangle
}

func (sp spec) solve(size int) (loc image.Point, level int) {
	for pt := sp.Min; pt.Y < sp.Max.Y-size; pt.Y++ {
		for pt.X = sp.Min.X; pt.X < sp.Max.X-size; pt.X++ {
			total := 0
			for dy := 0; dy < size; dy++ {
				for dx := 0; dx < size; dx++ {
					total += powerLevel(pt.Add(image.Pt(dx, dy)), sp.serial)
				}
			}
			if level < total {
				loc, level = pt, total
			}
		}
	}
	return loc, level
}

type partialSums struct {
	fuelGrid
	s []int
}

func buildPartialSums(serial int, bounds image.Rectangle) (ps partialSums) {
	ps.fuelGrid = buildFuelGrid(serial, bounds)
	ps.s = make([]int, len(ps.d))
	ps.accumulate()
	return ps
}

func (ps partialSums) accumulate() {
	for pt := ps.Min; pt.Y < ps.Max.Y; pt.Y++ {
		for pt.X = ps.Min.X; pt.X < ps.Max.X; pt.X++ {
			i, _ := ps.Index(pt)
			s := ps.d[i]                          // = cell value
			s += ps.sum(pt.Add(image.Pt(0, -1)))  // + up cum-cell value
			s += ps.sum(pt.Add(image.Pt(-1, 0)))  // + left cum-cell value
			s -= ps.sum(pt.Add(image.Pt(-1, -1))) // - up,left cum-cell value
			ps.s[i] = s
		}
	}
}

func (ps partialSums) sum(pt image.Point) int {
	if i, ok := ps.Index(pt); ok {
		return ps.s[i]
	}
	return 0
}

func (ps partialSums) solve(size int) (loc image.Point, level int) {
	for pt := ps.Min; pt.Y < ps.Max.Y-size; pt.Y++ {
		for pt.X = ps.Min.X; pt.X < ps.Max.X-size; pt.X++ {
			/*  |  |
			 * A| B|
			 *--+--+
			 *  |  |
			 * C| D|
			 *--+--+
			 */
			total := ps.sum(pt.Add(image.Pt(size, size))) // = lower-right (A + B + C + D)
			total -= ps.sum(pt.Add(image.Pt(size, 0)))    // - upper-right (A + B)
			total -= ps.sum(pt.Add(image.Pt(0, size)))    // - lower-left  (A + C)
			total += ps.sum(pt)                           // + upper-left  (A)
			// = (A + B + C + D) - (A + B) - (A + C) + A
			// = A + B + C + D - A - B - A - C + A
			// = A - A + B - B + C - C + D + A - A
			// = D
			if level < total {
				loc, level = pt.Add(image.Pt(1, 1)), total
			}
		}
	}
	return loc, level
}

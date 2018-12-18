package main

import (
	"flag"
	"image"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/jcorbin/anansi"
)

/* Ported from [antirez's LOLWUT](http://antirez.com/news/123)
 *
 * Creates output like:
 *
 * ⠀⡤⠤⠤⠤⠤⠤⠤⠤⠤⡤⠤⠤⠤⠤⠤⠤⠤⠤⡤⠤⠤⠤⠤⠤⠤⠤⠤⡄⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡏⠉⠉⠉⠉⠉⠉⠉⠉⡏⠉⠉⠉⠉⠉⠉⠉⠉⡏⠉⠉⠉⠉⠉⠉⠉⠉⡇⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠀⡷⠦⠤⢤⣤⣤⠤⠤⠤⡧⠤⠤⠤⠤⠤⠤⠤⢤⣧⣤⠤⠤⠤⠤⠴⠶⢶⠇⠀
 * ⢸⠀⠀⠀⠀⠀⠀⠉⠉⠒⡇⠀⠀⠀⠀⠀⠀⠀⢸⡇⠀⠀⠀⠀⠀⠀⠀⠘⡄⠀
 * ⡇⠀⠀⠀⠀⠀⠀⠀⠀⢸⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀
 * ⠁⠀⠀⠀⠀⠀⠀⠀⠀⡇⡇⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⢇⠀
 * ⡀⠀⠀⠀⠀⣀⡠⠔⢶⠁⡇⣀⣀⠤⠤⠔⠒⠊⠹⣷⡰⠢⠤⣀⡀⠀⢀⣀⣸⠀
 * ⢈⡩⠵⠒⠛⠤⠤⣀⡎⡾⡉⠉⠉⠉⠉⠉⠉⠉⠉⢿⠓⠒⠉⠉⠉⠉⠓⠢⠤⣀
 * ⡅⠀⠀⠀⠀⠀⠀⠀⠀⠘⣇⠀⠀⠀⠀⠀⠀⠀⠀⣾⠀⠀⠀⠀⠀⠀⠀⠀⠀⡎
 * ⠸⡀⠀⠀⠀⠀⠀⠀⠀⠀⢸⡀⠀⠀⠀⠀⠀⠀⡸⠀⡇⢀⢄⡀⠀⠀⠀⠀⡸⠀
 * ⠀⠱⡀⠀⠀⠀⠀⠀⠀⠀⣀⡷⠀⠀⠀⠀⠀⢠⣃⡠⡼⠊⠀⠈⠢⢄⠀⢠⠃⠀
 * ⠀⠀⢣⠀⠀⣀⣠⠴⡊⠉⠀⠸⢤⠶⢖⡉⠉⠁⢈⠝⠓⠢⠤⣀⡀⠀⠑⡮⡀⠀
 * ⠀⠀⣀⠷⠛⠉⠀⠀⠱⡀⠀⢀⠎⠀⠀⠈⠑⡢⢎⡀⠀⠀⠀⠀⠈⠉⠚⠀⠈⡕
 * ⠒⠉⠀⠀⠀⠀⠀⠀⠀⠱⣀⠎⠀⠀⠀⠀⠐⠥⡀⠈⢱⠂⠀⠀⠀⠀⠀⢠⠊⠀
 * ⢣⠀⠀⠀⠀⠀⠀⠀⠀⢀⠿⡀⠀⠀⠀⠀⠀⠀⠈⢲⢇⠀⠀⠀⠀⠀⡔⠁⠀⠀
 * ⠀⢣⠀⠀⠀⠀⠀⠀⠀⠚⢤⣱⠀⠀⠀⠀⠀⠀⢠⠃⠀⠉⠢⣀⢠⠊⠀⠀⠀⠀
 * ⠀⠀⠃⠀⠀⠀⠀⠀⠐⠊⠁⠀⠉⠒⠀⠀⠀⠀⠃⠀⠀⠀⠀⠀⠁⠀⠀⠀⠀⠀
 */

var sd schotterDemo

func main() {
	interactive := flag.Bool("i", false, "interactive mode")
	flag.Parse()
	var (
		// parameter  = parseArg(arg,         "name", default, min, max)
		cols          = parseArg(flag.Arg(0), "cols", 60, 1)
		squaresPerRow = parseArg(flag.Arg(1), "squares-per-row", 8, 1)
		squaresPerCol = parseArg(flag.Arg(2), "squares-per-col", 12, 1)
	)

	if *interactive {
		runInteractive() // TODO pass squaresPerRow / squaresPerCol ?
		return
	}

	sd.setup(cols, squaresPerRow, squaresPerCol)
	sd.draw()

	_, err := anansi.WriteBitmap(os.Stdout, sd.canvas)
	if err == nil {
		_, err = os.Stdout.WriteString("\n")
	}
	if err != nil {
		log.Fatalln(err)
	}
}

type schotterDemo struct {
	// config
	squaresPerRow int
	squaresPerCol int
	squareSide    int
	padding       int
	seed          int64
	angleOffset   float64

	// state
	canvas *anansi.Bitmap
	rand   *rand.Rand
}

// setup for the static demo, by computing config from command line arguments
// and allocating a canvas.
func (sd *schotterDemo) setup(cols, squaresPerRow, squaresPerCol int) {
	sd.squaresPerRow = squaresPerRow
	sd.squaresPerCol = squaresPerCol
	canvasWidth := cols * 2
	sd.padding = 0
	if canvasWidth > 4 {
		sd.padding = 2
	}
	sd.squareSide = int(float64(canvasWidth-sd.padding*2) / float64(sd.squaresPerRow))
	canvasHeight := sd.squareSide*sd.squaresPerCol + sd.padding*2
	sd.canvas = anansi.NewBitmapSize(image.Pt(canvasWidth, canvasHeight))
}

// draw a computer graphic art piece generated by Georg Nees in the 60s.  It
// explores the relationship between chaos and order.
func (sd *schotterDemo) draw() {
	sd.rand = rand.New(rand.NewSource(sd.seed))
	for y := 0; y < sd.squaresPerCol; y++ {
		for x := 0; x < sd.squaresPerRow; x++ {
			sx := x*sd.squareSide + sd.squareSide/2 + sd.padding
			sy := y*sd.squareSide + sd.squareSide/2 + sd.padding

			// Rotate and translate randomly as we go down to lower rows.
			angle := sd.angleOffset

			if y > 0 {
				r1 := sd.rand.Float64() / float64(sd.squaresPerCol) * float64(y)
				r2 := sd.rand.Float64() / float64(sd.squaresPerCol) * float64(y)
				r3 := sd.rand.Float64() / float64(sd.squaresPerCol) * float64(y)
				if sd.rand.Intn(2) == 1 {
					r1 = -r1
				}
				if sd.rand.Intn(2) == 1 {
					r2 = -r2
				}
				if sd.rand.Intn(2) == 1 {
					r3 = -r3
				}
				angle = sd.angleOffset + r1
				sx += int(r2 * float64(sd.squareSide) / 3)
				sy += int(r3 * float64(sd.squareSide) / 3)
			}

			drawSquare(sd.canvas, image.Pt(sx, sy), sd.squareSide, angle)
		}
	}
}

// drawSquare draws a square centered at the specified x,y coordinates, with
// the specified rotation angle and size.
func drawSquare(canvas *anansi.Bitmap, at image.Point, size int, angle float64) {
	// In order to write a rotated square, we use the trivial fact that the
	// parametric equation:
	//	 x, y = sin(k), cos(k)
	//
	// Describes a circle for values going from 0 to 2*PI. So basically if we
	// start at 45 degrees, that is k = PI/4, with the first point, and then we
	// find the other three points incrementing K by PI/2 (90 degrees), we'll
	// have the points of the square. In order to rotate the square, we just
	// start with k = PI/4 + rotation_angle, and we are done.
	//
	// Of course the vanilla equations above will describe the square inside a
	// circle of radius 1, so in order to draw larger squares we'll have to
	// multiply the obtained coordinates, and then translate them. However this
	// is much simpler than implementing the abstract concept of 2D shape and
	// then performing the rotation/translation transformation, so for LOLWUT
	// it's a good approach.

	// Adjust the desired size according to the fact that the square inscribed
	// into a circle of radius 1 has the side of length SQRT(2). This way
	// size becomes a simple multiplication factor we can use with our
	// coordinates to magnify them.
	fsize := math.Round(float64(size) / math.Sqrt2)

	// Draw the square.
	k := math.Pi/4 + angle
	last := rotPt(k, fsize).Add(at)
	for i := 0; i < 4; i++ {
		k += math.Pi / 2
		pt := rotPt(k, fsize).Add(at)
		drawLine(canvas, last, pt, true)
		last = pt
	}
}

func rotPt(angle, scale float64) image.Point {
	return image.Pt(
		int(math.Round(math.Sin(angle)*scale)),
		int(math.Round(math.Cos(angle)*scale)),
	)
}

// drawLine draws a bitmap line starting at the from point, through the to
// point using the Bresenham algorithm.
func drawLine(canvas *anansi.Bitmap, from, to image.Point, val bool) {
	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	sx, sy := 1, 1
	if from.X >= to.X {
		sx = -1
	}
	if from.Y >= to.Y {
		sy = -1
	}
	err := dx - dy
	canvas.Set(from, val)
	for from != to {
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			from.X += sx
		}
		if e2 < dx {
			err += dx
			from.Y += sy
		}
		canvas.Set(from, val)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func parseArg(arg, name string, def, min int) int {
	if arg == "" {
		return def
	}
	n, err := strconv.Atoi(arg)
	if err != nil {
		log.Fatalf("invalid %s argument %q: %v", name, arg, err)
	}
	if n < min {
		return min
	}
	return n
}

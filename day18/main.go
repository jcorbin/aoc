package main

import (
	"bufio"
	"flag"
	"image"
	"io"
	"log"
	"math"
	"os"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/aoc/internal/geom"
)

var numRounds = flag.Int("n", 0, "num rounds")

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type board struct {
	geom.RCore
	d []byte
}

func run(in, out *os.File) error {
	brd, err := read(in)
	if err != nil {
		return err
	}

	stencil := [8]image.Point{
		image.Pt(-1, -1), image.Pt(0, -1), image.Pt(1, -1),
		image.Pt(-1, 0) /*image.Pt(0, 0),*/, image.Pt(1, 0),
		image.Pt(-1, 1), image.Pt(0, 1), image.Pt(1, 1),
	}

	countem := func(p image.Point) (o, t, l int) {
		for _, dp := range stencil {
			if i, ok := brd.Index(p.Add(dp)); ok {
				switch brd.d[i] {
				case '.':
					o++
				case '|':
					t++
				case '#':
					l++
				}
			}
		}
		return o, t, l
	}

	// part 1
	tmp := brd
	tmp.d = make([]byte, len(tmp.d))
	for i := 0; i < *numRounds; i++ {
		to, tt, tl := 0, 0, 0

		for pt := tmp.Min; pt.Y < tmp.Max.Y; pt.Y++ {
			for pt.X = tmp.Min.X; pt.X < tmp.Max.X; pt.X++ {
				i, _ := brd.Index(pt)
				_, t, l := countem(pt)

				switch brd.d[i] {

				// An *open* acre will become filled with *trees* if *three or
				// more* adjacent acres contained trees. Otherwise, nothing
				// happens.
				case '.':
					if t >= 3 {
						tmp.d[i] = '|'
						tt++
					} else {
						tmp.d[i] = '.'
						to++
					}

					// An acre filled with *trees* will become a *lumberyard* if *three
					// or more* adjacent acres were lumberyards. Otherwise, nothing
					// happens.
				case '|':
					if l >= 3 {
						tmp.d[i] = '#'
						tl++
					} else {
						tmp.d[i] = '|'
						tt++
					}

					// An acre containing a *lumberyard* will remain a *lumberyard* if
					// it was adjacent to *at least one other lumberyard and at least
					// one acre containing trees*. Otherwise, it becomes *open*.
				case '#':
					if l >= 1 && t >= 1 {
						tmp.d[i] = '#'
						tl++
					} else {
						tmp.d[i] = '.'
						to++
					}

				}
			}
		}
		brd.d, tmp.d = tmp.d, brd.d

		// log.Printf("open:%v trees:%v lumberyards:%v", to, tt, tl)
		log.Printf("[%v]: %v", i, tt*tl)
	}

	// part 2
	// TODO

	return nil
}

func read(r io.Reader) (brd board, _ error) {
	sc := bufio.NewScanner(r)
	// sc.Split(bufio.ScanWords)
	for sc.Scan() {
		line := sc.Text()
		for i := 0; i < len(line); i++ {
			brd.d = append(brd.d, line[i])
		}
	}

	n := int(math.Ceil(math.Sqrt(float64(len(brd.d)))))
	brd.Stride = n
	brd.Max.X = n
	brd.Max.Y = n

	return brd, sc.Err()
}

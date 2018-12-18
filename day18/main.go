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

func (brd board) step(tmp []byte) (to, tt, tl int) {
	for pt := brd.Min; pt.Y < brd.Max.Y; pt.Y++ {
		for pt.X = brd.Min.X; pt.X < brd.Max.X; pt.X++ {
			i, ok := brd.Index(pt)
			if !ok {
				continue
			}

			t, l := 0, 0
			for _, dp := range [8]image.Point{
				image.Pt(-1, -1), image.Pt(0, -1), image.Pt(1, -1),
				image.Pt(-1, 0) /*image.Pt(0, 0),*/, image.Pt(1, 0),
				image.Pt(-1, 1), image.Pt(0, 1), image.Pt(1, 1),
			} {
				if i, ok := brd.Index(pt.Add(dp)); ok {
					switch brd.d[i] {
					case '|':
						t++
					case '#':
						l++
					}
				}
			}

			switch brd.d[i] {

			// An *open* acre will become filled with *trees* if *three or
			// more* adjacent acres contained trees. Otherwise, nothing
			// happens.
			case '.':
				if t >= 3 {
					tmp[i] = '|'
					tt++
				} else {
					tmp[i] = '.'
					to++
				}

			// An acre filled with *trees* will become a *lumberyard* if *three
			// or more* adjacent acres were lumberyards. Otherwise, nothing
			// happens.
			case '|':
				if l >= 3 {
					tmp[i] = '#'
					tl++
				} else {
					tmp[i] = '|'
					tt++
				}

			// An acre containing a *lumberyard* will remain a *lumberyard* if
			// it was adjacent to *at least one other lumberyard and at least
			// one acre containing trees*. Otherwise, it becomes *open*.
			case '#':
				if l >= 1 && t >= 1 {
					tmp[i] = '#'
					tl++
				} else {
					tmp[i] = '.'
					to++
				}

			}
		}
	}
	return to, tt, tl
}

func run(in, out *os.File) error {
	brd, err := read(in)
	if err != nil {
		return err
	}
	tmp := make([]byte, len(brd.d))
	for i := 0; i < *numRounds; i++ {
		_, tt, tl := brd.step(tmp)
		brd.d, tmp = tmp, brd.d
		log.Printf("[%v]: %v", i, tt*tl)
	}
	return nil
}

func read(r io.Reader) (brd board, _ error) {
	sc := bufio.NewScanner(r)
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

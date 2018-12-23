package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type point3 struct {
	X, Y, Z int
}

func (p point3) sub(other point3) point3 {
	p.X -= other.X
	p.Y -= other.Y
	p.Z -= other.Z
	return p
}

func (p point3) abs() point3 {
	if p.X < 0 {
		p.X = -p.X
	}
	if p.Y < 0 {
		p.Y = -p.Y
	}
	if p.Z < 0 {
		p.Z = -p.Z
	}
	return p
}

func (p point3) sum() int {
	return p.X + p.Y + p.Z
}

func (p point3) add(other point3) point3 {
	p.X += other.X
	p.Y += other.Y
	p.Z += other.Z
	return p
}

func (p point3) eachWithin(r int, each func(point3, int)) {
	been := make(map[point3]struct{}, 6*r)
	for _, dp := range [...]point3{
		{-1, 0, 0}, {1, 0, 0},
		{0, -1, 0}, {0, 1, 0},
		{0, 0, -1}, {0, 0, 1},
	} {
		np := p.add(dp)
		if _, have := been[np]; have {
			continue
		}
		been[np] = struct{}{}
		each(np, r)
		if r > 1 {
			np.eachWithin(r-1, each)
		}
	}
}

type bot struct {
	point3
	R int
}

func run(in, out *os.File) error {
	bots, err := read(in)
	if err != nil {
		return err
	}

	type idSet map[int]struct{}

	// space := make(map[point3]idSet, 1024)
	// p := b.point3
	// ids := space[p]
	// if ids == nil {
	// 	ids = make(idSet, len(bots))
	// 	space[p] = ids
	// }
	// ids[id] = struct{}{}

	// part 1
	var best, bestRadius int

	for id, b := range bots {
		if bestRadius < b.R {
			best, bestRadius = id, b.R
		}
	}
	log.Printf("best #%v %v", best, bots[best])

	ids := make(idSet, len(bots))
	bestBot := bots[best]
	for id, b := range bots {
		d := b.point3.sub(bestBot.point3).abs().sum()
		if d <= bestBot.R {
			ids[id] = struct{}{}
		}
	}

	// bestBot.point3.eachWithin(bestBot.R, func(p point3, rem int) {
	// 	for id := range space[p] {
	// 		ids[id] = struct{}{}
	// 	}
	// })

	log.Printf("%v in range:", len(ids))

	// for id := range ids {
	// 	b := bots[id]
	// 	d := b.point3.sub(bestBot.point3).abs().sum()
	// 	log.Printf("#%v in range d:%v bot:%v", id, d, b)
	// }

	// part 2
	// TODO

	return nil
}

var botPattern = regexp.MustCompile(`^pos=<(-?\d+),(-?\d+),(-?\d+)>, r=(-?\d+)$`)

func read(r io.Reader) (bots []bot, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := botPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			return nil, fmt.Errorf("unrecognized line %q, expected %v", line, botPattern)
		}

		var b bot
		b.X, _ = strconv.Atoi(parts[1])
		b.Y, _ = strconv.Atoi(parts[2])
		b.Z, _ = strconv.Atoi(parts[3])
		b.R, _ = strconv.Atoi(parts[4])

		bots = append(bots, b)
	}
	return bots, sc.Err()
}

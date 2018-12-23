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

var verbose = flag.Bool("v", false, "verbose logging")

func main() {
	flag.Parse()

	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type bot struct {
	point3
	R int
}

func (b bot) bounds() cube {
	return cube{
		b.point3.subN(b.R),
		b.point3.addN(b.R + 1),
	}
}

func run(in, out *os.File) error {
	bots, err := read(in)
	if err != nil {
		return err
	}

	type idSet map[int]struct{}

	// part 1
	var best, bestRadius int
	for id, b := range bots {
		if bestRadius < b.R {
			best, bestRadius = id, b.R
		}
	}
	if *verbose {
		log.Printf("best #%v %v", best, bots[best])
	}
	ids := make(idSet, len(bots))
	bestBot := bots[best]
	for id, b := range bots {
		d := b.point3.sub(bestBot.point3).abs().sum()
		if d <= bestBot.R {
			ids[id] = struct{}{}
		}
	}
	log.Printf("%v in range:", len(ids))

	// part 2
	bestSpot, bestN := search(bots)
	log.Printf("best @%v dist:%v n:%v", bestSpot, bestSpot.abs().sum(), bestN)

	return nil
}

func search(bots []bot) (point3, int) {
	bnd := bounds(bots)
	dist := 1
	for dx := bnd.dx(); dist < dx; {
		dist *= 2
	}
	for {
		if *verbose {
			log.Printf("searching in %v dist:%v", bnd, dist)
		}
		best, bestN := findBest(bots, dist, bnd)
		if dist == 1 {
			return best, bestN
		}
		bnd.min.X, bnd.max.X = best.X-dist, best.X+dist
		bnd.min.Y, bnd.max.Y = best.Y-dist, best.Y+dist
		bnd.min.Z, bnd.max.Z = best.Z-dist, best.Z+dist
		dist /= 2
	}
}

func findBest(bots []bot, dist int, bnd cube) (best point3, bestN int) {
	for p := bnd.min; p.Z < bnd.max.Z; p.Z += dist {
		for p.Y = bnd.min.Y; p.Y < bnd.max.Y; p.Y += dist {
			for p.X = bnd.min.X; p.X < bnd.max.X; p.X += dist {
				n := 0
				for _, b := range bots {
					if d := p.sub(b.point3).abs().sum(); (d-b.R)/dist <= 0 {
						n++
					}
				}
				if n < bestN {
					continue
				} else if n > bestN {
					best, bestN = p, n
				} else if p.abs().sum() < best.abs().sum() {
					best, bestN = p, n
				}
			}
		}
	}
	return best, bestN
}

func bounds(bots []bot) (bnd cube) {
	for i, b := range bots {
		c := b.point3.toCube()
		if i == 0 {
			bnd = c
		} else {
			bnd = bnd.union(c)
		}
	}
	return bnd
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

type point3 struct{ X, Y, Z int }
type cube struct{ min, max point3 }

var zc cube

func (c cube) dx() int { return c.max.X - c.min.X }
func (c cube) dy() int { return c.max.Y - c.min.Y }
func (c cube) dz() int { return c.max.Z - c.min.Z }

func (c cube) size() point3 {
	return c.max.sub(c.min)
}

func (c cube) empty() bool {
	return c.min.X >= c.max.X ||
		c.min.Y >= c.max.Y ||
		c.min.Z >= c.max.Z
}

func (c cube) intersect(d cube) cube {
	if c.min.X < d.min.X {
		c.min.X = d.min.X
	}
	if c.min.Y < d.min.Y {
		c.min.Y = d.min.Y
	}
	if c.min.Z < d.min.Z {
		c.min.Z = d.min.Z
	}
	if c.max.X > d.max.X {
		c.max.X = d.max.X
	}
	if c.max.Y > d.max.Y {
		c.max.Y = d.max.Y
	}
	if c.max.Z > d.max.Z {
		c.max.Z = d.max.Z
	}
	if c.empty() {
		return zc
	}
	return c
}

func (c cube) union(d cube) cube {
	if c.empty() {
		return d
	}
	if d.empty() {
		return c
	}
	if c.min.X > d.min.X {
		c.min.X = d.min.X
	}
	if c.min.Y > d.min.Y {
		c.min.Y = d.min.Y
	}
	if c.min.Z > d.min.Z {
		c.min.Z = d.min.Z
	}
	if c.max.X < d.max.X {
		c.max.X = d.max.X
	}
	if c.max.Y < d.max.Y {
		c.max.Y = d.max.Y
	}
	if c.max.Z < d.max.Z {
		c.max.Z = d.max.Z
	}
	return c
}

func (p point3) toCube() cube {
	return cube{p, p.addN(1)}
}

func (p point3) in(c cube) bool {
	return c.min.X <= p.X && p.X < c.max.X &&
		c.min.Y <= p.Y && p.Y < c.max.Y &&
		c.min.Z <= p.Z && p.Z < c.max.Z
}

func (p point3) subN(n int) point3 {
	p.X -= n
	p.Y -= n
	p.Z -= n
	return p
}

func (p point3) addN(n int) point3 {
	p.X += n
	p.Y += n
	p.Z += n
	return p
}

func (p point3) sub(other point3) point3 {
	p.X -= other.X
	p.Y -= other.Y
	p.Z -= other.Z
	return p
}

func (p point3) add(other point3) point3 {
	p.X += other.X
	p.Y += other.Y
	p.Z += other.Z
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

func (p point3) div(n int) point3 {
	p.X /= n
	p.Y /= n
	p.Z /= n
	return p
}

func (p point3) mul(n int) point3 {
	p.X *= n
	p.Y *= n
	p.Z *= n
	return p
}

func (p point3) eachWithin(r int, each func(point3)) {
	min := p.subN(r)
	max := p.addN(r + 1)
	for np := min; np.Z < max.Z; np.Z++ {
		for np.Y = min.Y; np.Y < max.Y; np.Y++ {
			for np.X = min.X; np.X < max.X; np.X++ {
				d := np.sub(p).abs().sum()
				if d <= r {
					each(np)
				}
			}
		}
	}
}

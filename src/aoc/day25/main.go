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

var radius = flag.Int("r", 3, "clustering radius")

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

func run(in, out *os.File) error {
	pts, err := readPoints(in)
	if err != nil {
		return err
	}

	log.Printf("read points: %v", pts)
	_, numClusters := clusterPoints(pts, *radius)
	log.Printf("grouped into %v clusters", numClusters)

	return nil
}

func clusterPoints(pts []point4, r int) (clusterIDs []int, numClusters int) {
	dirs := make([]point4, 0, 16*r*r*r*r)
	for dx := -r; dx <= r; dx++ {
		for dy := -r; dy <= r; dy++ {
			for dz := -r; dz <= r; dz++ {
				for dw := -r; dw <= r; dw++ {
					dir := point4{dx, dy, dz, dw}
					if d := dir.abs().sum(); d > 0 && d <= r {
						dirs = append(dirs, dir)
					}
				}
			}
		}
	}

	// setup point -> id reverse index
	pt2id := make(map[point4]int, 2*len(pts))
	for i, pt := range pts {
		id := i + 1
		pt2id[pt] = id
	}

	// perform union-find of each point to any r-neighbor
	var uf unionFind
	uf.init(len(pts))
	for i, pt := range pts {
		id := i + 1
		for _, dir := range dirs {
			if nid := pt2id[pt.add(dir)]; nid != 0 {
				uf.union(id-1, nid-1)
			}
		}
	}

	return uf.id, uf.count
}

type pointID struct {
	point4
	id int
}

var point4Pattern = regexp.MustCompile(`^(-?\d+),(-?\d+),(-?\d+),(-?\d+)$`)

func readPoints(r io.Reader) (pts []point4, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := point4Pattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			return nil, fmt.Errorf("bad line %q, expected %v", line, point4Pattern)
		}

		var pt point4
		pt.X, _ = strconv.Atoi(parts[1])
		pt.Y, _ = strconv.Atoi(parts[2])
		pt.Z, _ = strconv.Atoi(parts[3])
		pt.W, _ = strconv.Atoi(parts[4])
		pts = append(pts, pt)
	}
	return pts, sc.Err()
}

type point4 struct{ X, Y, Z, W int }

func (p point4) addN(n int) point4 {
	p.X += n
	p.Y += n
	p.Z += n
	p.W += n
	return p
}

func (p point4) add(other point4) point4 {
	p.X += other.X
	p.Y += other.Y
	p.Z += other.Z
	p.W += other.W
	return p
}

func (p point4) abs() point4 {
	if p.X < 0 {
		p.X = -p.X
	}
	if p.Y < 0 {
		p.Y = -p.Y
	}
	if p.Z < 0 {
		p.Z = -p.Z
	}
	if p.W < 0 {
		p.W = -p.W
	}
	return p
}

func (p point4) sum() int { return p.X + p.Y + p.Z + p.W }

type unionFind struct {
	id    []int
	rank  []int
	count int
}

func (uf *unionFind) init(n int) {
	uf.id = make([]int, n)
	uf.rank = make([]int, n)
	uf.count = n
	for i := 0; i < n; i++ {
		uf.id[i] = i
	}
}

func (uf *unionFind) find(p int) int {
	for p != uf.id[p] {
		uf.id[p] = uf.id[uf.id[p]] // Path compression using halving.
		p = uf.id[p]
	}
	return p
}

func (uf *unionFind) union(p, q int) {
	i, j := uf.find(p), uf.find(q)
	if i != j {
		uf.count--
		if uf.rank[i] < uf.rank[j] {
			uf.id[i] = j
		} else if uf.rank[i] > uf.rank[j] {
			uf.id[j] = i
		} else {
			uf.id[j] = i
			uf.rank[i]++
		}
	}
}

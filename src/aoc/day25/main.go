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
	clusters := clusterPoints(pts, *radius)
	log.Printf("grouped into %v clusters", len(clusters))

	return nil
}

func clusterPoints(pts []point4, r int) (clusters [][]point4) {
	for j := 0; j < len(pts); j++ {
		pt := pts[j]

		// strike clusters that point is close to
		var coll []point4
		for i, cluster := range clusters {
			for _, opt := range cluster {
				if d := opt.sub(pt).abs().sum(); d <= r {
					coll = append(coll, cluster...)
					clusters[i] = nil
					break
				}
			}
		}

		// compact struck clusters
		if len(coll) > 0 {
			i := 0
			for j := 0; j < len(clusters); j++ {
				if clusters[j] != nil {
					clusters[i] = clusters[j]
					i++
				}
			}
			clusters = clusters[:i]
		}

		// append point, and new cluster
		clusters = append(clusters, append(coll, pt))
	}
	return clusters
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

func (p point4) sub(other point4) point4 {
	p.X -= other.X
	p.Y -= other.Y
	p.Z -= other.Z
	p.W -= other.W
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

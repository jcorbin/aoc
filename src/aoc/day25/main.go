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
	// only goes up to track unique clusters; numClusters can go back down
	// after coalescing.
	nextClusterID := 0

	// assigning one cluster id to every point given
	clusterIDs = make([]int, len(pts))

	for j := 0; j < len(pts); j++ {
		clusterIDs[j] = func(pt point4) int {

			// find a prior cluster
			for i := 0; i < j; i++ {
				if pts[i].sub(pt).abs().sum() <= r {
					clusterID := clusterIDs[i]

					// coalesce any other clusters...
					for ; i < j; i++ {
						if pts[i].sub(pt).abs().sum() <= r {
							if ocid := clusterIDs[i]; ocid != clusterID {

								// ...rewriting all their prior assignments
								clusterIDs[i] = clusterID
								for k := 0; k < j; k++ {
									if clusterIDs[k] == ocid {
										clusterIDs[k] = clusterID
									}
								}
								numClusters--
							}
						}
					}
					return clusterID
				}
			}

			// assign a new cluster if we didn't find a prior
			numClusters++
			nextClusterID++
			return nextClusterID

		}(pts[j])
	}
	return clusterIDs, numClusters
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

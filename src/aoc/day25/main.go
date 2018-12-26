package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
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
	// order points in z-order; builds a linear quadtree for range queries.
	sort.Sort(zSort(pts))

	// only goes up to track unique clusters; numClusters can go back down
	// after coalescing.
	nextClusterID := 0

	// assigning one cluster id to every point given
	clusterIDs = make([]int, len(pts))

	for j := 0; j < len(pts); j++ {
		clusterIDs[j] = func(pt point4) int {

			// combine with an prior assignments within radius (only possible in z-region)
			i := 0
			zq := pt.addN(-r) // query for minimum of possible z-region
			for ; i < j && pts[i].zless(zq); i++ {
			}
			zq = pt.addN(r + 1) // query for maximum of possible z-region
			for ; i < j && pts[i].zless(zq); i++ {
				if pts[i].sub(pt).abs().sum() <= r {
					clusterID := clusterIDs[i]

					// coalesce any other clusters...
					for i++; i < j && pts[i].zless(zq); i++ {
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

type zSort []point4

func (zs zSort) Len() int               { return len(zs) }
func (zs zSort) Less(i int, j int) bool { return zs[i].zless(zs[j]) }
func (zs zSort) Swap(i int, j int)      { zs[i], zs[j] = zs[j], zs[i] }

func (p point4) zless(other point4) (r bool) {
	x := 0

	if y := p.X ^ other.X; lessMSB(x, y) {
		x = y
		r = p.X < other.X
	}

	if y := p.Y ^ other.Y; lessMSB(x, y) {
		x = y
		r = p.Y < other.Y
	}

	if y := p.Z ^ other.Z; lessMSB(x, y) {
		x = y
		r = p.Z < other.Z
	}

	if y := p.W ^ other.W; lessMSB(x, y) {
		// x = y
		r = p.W < other.W
	}

	return r
}

func lessMSB(x, y int) bool { return x < y && x < (x^y) }

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

package main

import (
	"bufio"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

func main() {
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

type claim struct {
	id int
	image.Rectangle
}

func run(r io.Reader) error {
	claims, err := readClaims(r)
	if err != nil {
		return err
	}

	log.Printf("read %v claims", len(claims))

	max := image.ZP
	for _, c := range claims {
		if max.X < c.Max.X {
			max.X = c.Max.X
		}
		if max.Y < c.Max.Y {
			max.Y = c.Max.Y
		}
	}
	log.Printf("max: %v", max)

	counts := make([]int, max.X*max.Y)
	for _, c := range claims {
		pt := c.Min
		for ; pt.Y < c.Max.Y; pt.Y++ {
			pt.X = c.Min.X
			i := max.X*pt.Y + pt.X
			for ; pt.X < c.Max.X; pt.X++ {
				counts[i]++
				i++
			}
		}
	}

	// var bi anansi.Bitmap
	// bi.Stride = max.X
	// bi.Bit = make([]bool, len(counts))
	// bi.Rect.Max = max
	// for i := 0; i < len(counts); i++ {
	// 	bi.Bit[i] = counts[i] > 1
	// }
	// anansi.WriteBitmap(os.Stdout, &bi)

	n := 0
	for i := 0; i < len(counts); i++ {
		if counts[i] > 1 {
			n++
		}
	}
	log.Printf("%v squares overlap", n)

	for _, c := range claims {
		if func() bool {
			pt := c.Min
			for ; pt.Y < c.Max.Y; pt.Y++ {
				pt.X = c.Min.X
				i := max.X*pt.Y + pt.X
				for ; pt.X < c.Max.X; pt.X++ {
					if counts[i] != 1 {
						return false
					}
					i++
				}
			}
			return true
		}() {
			log.Printf("claim %#v is clean", c)
		}
	}

	return nil
}

var claimPattern = regexp.MustCompile(
	`^#(\d+) +@ +(\d+),(\d+): +(\d+)x(\d+)$`,
)

func readClaims(r io.Reader) ([]claim, error) {
	var claims []claim
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := claimPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			log.Printf("NO MATCH in %q", line)
			continue
		}

		var c claim
		c.id, _ = strconv.Atoi(parts[1])
		c.Min.X, _ = strconv.Atoi(parts[2])
		c.Min.Y, _ = strconv.Atoi(parts[3])
		c.Max.X, _ = strconv.Atoi(parts[4])
		c.Max.Y, _ = strconv.Atoi(parts[5])
		c.Max.X += c.Min.X
		c.Max.Y += c.Min.Y
		claims = append(claims, c)
	}
	return claims, sc.Err()
}

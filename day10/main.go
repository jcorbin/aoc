package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
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

type space struct {
	t    int
	p, v []image.Point
}

func (sp space) bounds() (r image.Rectangle) {
	r.Min = sp.p[0]
	r.Max = sp.p[0]
	for i := 1; i < len(sp.p); i++ {
		p := sp.p[i]
		if r.Min.X > p.X {
			r.Min.X = p.X
		}
		if r.Min.Y > p.Y {
			r.Min.Y = p.Y
		}
		if r.Max.X < p.X {
			r.Max.X = p.X
		}
		if r.Max.Y < p.Y {
			r.Max.Y = p.Y
		}
	}
	r.Max.X++
	r.Max.Y++
	return r
}

func (sp space) render() anansi.Bitmap {
	var bi anansi.Bitmap
	bnd := sp.bounds()

	size := bnd.Size()
	bi.Bit = make([]bool, size.X*size.Y)
	bi.Rect.Max = size
	bi.Stride = size.X
	for _, p := range sp.p {
		bi.Set(p.Sub(bnd.Min), true)
	}
	return bi
}

func (sp *space) update() {
	sp.t++
	for i := range sp.p {
		sp.p[i] = sp.p[i].Add(sp.v[i])
	}
}

func (sp *space) rewind() {
	sp.t--
	for i := range sp.p {
		sp.p[i] = sp.p[i].Sub(sp.v[i])
	}
}

func run(in, out *os.File) error {
	sp, err := read(in)
	if err != nil {
		return err
	}

	log.Printf("read: %v", len(sp.p))

	lastn := 0
	for {
		sz := sp.bounds().Size()
		n := sz.X * sz.Y
		var dn int
		if lastn != 0 {
			dn = n - lastn
			if dn >= 0 {
				break
			}
		}
		lastn = n
		sp.update()
		// log.Printf("tick %v n:%v dn:%v", sp.t, n, dn)
	}

	sp.rewind()

	fmt.Fprintf(out, "--- t:%v %v\r\n", sp.t, sp.bounds().Size())
	fmt.Fprintf(out, "\r\n")

	bi := sp.render()
	anansi.WriteBitmap(out, &bi)

	return nil
}

var linePat = regexp.MustCompile(
	`^position=< *(-?\d+), *(-?\d+)> +velocity=< *(-?\d+), *(-?\d+)>$`,
)

func read(r io.Reader) (sp space, err error) {
	sc := bufio.NewScanner(r)
	// sc.Split(bufio.ScanWords)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		parts := linePat.FindStringSubmatch(line)
		if len(parts) == 0 {
			log.Printf("no match %q", line)
			continue
		}

		i := len(sp.p)
		sp.p = append(sp.p, image.ZP)
		sp.v = append(sp.v, image.ZP)

		sp.p[i].X, err = strconv.Atoi(parts[1])
		if err != nil {
			return sp, fmt.Errorf("invalid pos.X %q: %v", parts[1], err)
		}
		sp.p[i].Y, err = strconv.Atoi(parts[2])
		if err != nil {
			return sp, fmt.Errorf("invalid pos.X %q: %v", parts[2], err)
		}

		sp.v[i].X, err = strconv.Atoi(parts[3])
		if err != nil {
			return sp, fmt.Errorf("invalid vel.X %q: %v", parts[3], err)
		}
		sp.v[i].Y, err = strconv.Atoi(parts[4])
		if err != nil {
			return sp, fmt.Errorf("invalid vel.X %q: %v", parts[4], err)
		}

	}
	return sp, sc.Err()
}

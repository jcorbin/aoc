package main

import (
	"aoc/internal/geom"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var justSolve = flag.Bool("solve", false, "just solve it")

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type space struct {
	t    int
	p, v []image.Point
}

func (sp space) bounds() (r image.Rectangle) {
	for i := 0; i < len(sp.p); i++ {
		p := sp.p[i]
		if i == 0 {
			r = geom.PointRect(p)
		} else {
			r = r.Union(geom.PointRect(p))
		}
	}
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

	if *justSolve || !anansi.IsTerminal(out) {
		return solve(sp, out)
	}

	if !anansi.IsTerminal(in) {
		f, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
		if err != nil {
			return err
		}
		in = f
		defer in.Close()
	}

	var haveInput anansi.InputSignal
	haveInput.Send("initial render")
	term := anansi.NewTerm(in, out,
		&haveInput,
	)
	term.SetRaw(true)
	term.SetEcho(false)
	term.AddMode(ansi.ModeAlternateScreen)

	seeking := false
	var size image.Point
	term.Set = ansi.ED.With(2).AppendTo(term.Set)
	term.Set = ansi.CUP.AppendTo(term.Set)

	return term.RunLoop(anansi.LoopClientFuncs(
		func(term *anansi.Term) (redraw bool, _ error) {
			var to <-chan time.Time
			if seeking {
				to = time.After(time.Millisecond)
			}
			select {

			case <-to:
				sp.update()
				if newSize := sp.bounds().Size(); newSize.X*newSize.Y > size.X*size.Y {
					seeking = false
				}
				return true, nil

			case <-haveInput.C:
				if _, err := term.ReadAny(); err != nil {
					return false, err
				}
				for e, _, ok := term.Decode(); ok; e, _, ok = term.Decode() {
					switch e {
					// Ctrl-C to quit
					case 0x03:
						return false, errors.New("goodbye")
					// arrow keys to advance/rewind
					case ansi.CUU:
						sp.rewind()
						redraw = true
					case ansi.CUD:
						sp.update()
						redraw = true
					}
				}
				return redraw, nil

			}
		},

		func(w io.Writer) (n int64, err error) {
			termSize, err := term.Size()
			if err != nil {
				return 0, err
			}

			size = sp.bounds().Size()
			m, err := fmt.Fprintf(w, "--- t:%v %v\r\n", sp.t, size)
			n += int64(m)
			if err != nil {
				return n, err
			}

			if seeking = size.X/2 > termSize.X || size.Y/4 > termSize.Y; seeking {
				return n, nil
			}

			bi := sp.render()
			m, err = anansi.WriteBitmap(w, &bi)
			n += int64(m)
			if err != nil {
				return n, err
			}
			m, err = fmt.Fprintf(w, "\r\n")
			n += int64(m)
			return n, err
		},
	))
}

func solve(sp space, out *os.File) error {
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
	bi := sp.render()
	anansi.WriteBitmap(out, &bi)
	fmt.Fprintf(out, "\r\n")

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

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/aoc/internal/quadindex"
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type cartType uint8

const (
	cart cartType = 1 << iota
	cartCrash
	cartTrack
)

type cartDirection uint8

const (
	cartDirUp    cartDirection = 0x00
	cartDirRight cartDirection = 0x01
	cartDirDown  cartDirection = 0x02
	cartDirLeft  cartDirection = 0x03

	cartTrackV cartDirection = 0x04
	cartTrackH cartDirection = 0x08
	cartTrackX cartDirection = 0x08 | 0x04
	cartTrackB cartDirection = 0x08 | 0x10
	cartTrackF cartDirection = 0x08 | 0x20
)

const (
	cartDirMask      cartDirection = 0x01 | 0x02
	cartTrackDirMask cartDirection = 0x04 | 0x08 | 0x10 | 0x20
)

var (
	// We need to handle these signals so that we restore terminal state
	// properly (raw mode and exit the alternate screen).
	halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)

	// terminal resize signals
	resize = anansi.Notify(syscall.SIGWINCH)

	// input availability notification
	inputReady anansi.InputSignal

	// The virtual screen that will be our canvas.
	screen anansi.Screen

	// in-memory log buffer
	logs bytes.Buffer
)

func init() {
	f, err := os.Create("cart_world.log")
	if err != nil {
		log.Fatalf("failed to create cart_world.log: %v", err)
	}
	log.SetOutput(io.MultiWriter(
		&logs,
		f,
	))
}

func overlayLogs() {
	n := bytes.Count(logs.Bytes(), []byte{'\n'})
	lb := logs.Bytes()
	for n > 5 {
		off := bytes.IndexByte(lb, '\n')
		if off < 0 {
			break
		}
		lb = lb[off+1:]
		logs.Next(off + 1)
		n--
	}

	screen.To(ansi.Pt(1, screen.Bounds().Dy()-n+1))

	lb = logs.Bytes()
	for {
		off := bytes.IndexByte(lb, '\n')
		if off < 0 {
			screen.Write(lb)
			break
		}
		screen.Write(lb[:off])
		screen.WriteString("\r\n")
		lb = lb[off+1:]
	}
}

type cartWorld struct {
	quadindex.Index
	b image.Rectangle
	p []image.Point
	t []cartType      // is cart? has track?
	d []cartDirection // direction of cart and/or track here
	s []int           // cart state

	carts []int

	timer *time.Timer
}

func run(in, out *os.File) error {
	var world cartWorld
	if err := world.load(in); err != nil {
		return err
	}

	in, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer in.Close()

	term := anansi.NewTerm(in, out,
		&halt,
		&resize,
		&inputReady,
		&screen,
	)
	term.SetRaw(true)
	term.AddMode(ansi.ModeAlternateScreen)
	resize.Send("initialize screen size")
	return term.RunWith(&world)
}

func (world *cartWorld) Run(term *anansi.Term) error {
	world.timer = time.NewTimer(100 * time.Second)
	world.timer.Stop()
	return term.Loop(world)
}

func (world *cartWorld) Update(term *anansi.Term) (redraw bool, _ error) {
	select {
	case sig := <-halt.C:
		return false, anansi.SigErr(sig)

	case <-resize.C:
		if err := screen.SizeToTerm(term); err != nil {
			return false, err
		}
		redraw = true

	case <-inputReady.C:
		_, err := term.ReadAny()
		update, herr := world.handleInput(term)
		if err == nil {
			err = herr
		}
		if err != nil {
			return false, err
		}
		if update {
			world.tick()
		}
		redraw = true

	case <-world.timer.C:
		world.tick()
		redraw = true
	}
	return redraw, nil
}

func (world *cartWorld) tick() {
	log.Printf("TODO tick")
	// for _, id := range world.carts {
	// 	TODO
	// }
	// Index.Update(i, p)
	// Index.Delete(i, p)
	// Index.At(p)
}

func (world *cartWorld) handleInput(term *anansi.Term) (update bool, _ error) {
	for e, _, ok := term.Decode(); ok; e, _, ok = term.Decode() {
		switch e {

		case 0x03: // stop on Ctrl-C
			return false, fmt.Errorf("read %v", e)

		case 0x0c: // clear screen on Ctrl-L
			screen.Clear()           // clear virtual contents
			screen.To(ansi.Pt(1, 1)) // cursor back to top
			screen.Invalidate()      // force full redraw

		// step
		case ansi.Escape('.'):
			update = true

			// TODO play/pause
			// case ansi.Escape(' '):
			// world.playing = true

			// TODO speed control

		}
	}

	return update, nil
}

func (world *cartWorld) WriteTo(w io.Writer) (n int64, err error) {
	screen.Clear()
	world.render(screen.Grid, image.ZP)
	overlayLogs()
	return screen.WriteTo(w)
}

func (world *cartWorld) render(g anansi.Grid, offset image.Point) {
	for id, t := range world.t {
		if id == 0 {
			continue
		}

		p := world.p[id].Add(offset)
		gi, ok := g.CellOffset(ansi.PtFromImage(p))
		if !ok {
			continue
		}

		var r rune
		var a ansi.SGRAttr
		switch {

		case t&cartCrash != 0:
			r = 'X'
			a |= ansi.RGB(192, 64, 64).FG()

		case t&cart != 0:
			switch world.d[id] & cartDirMask {
			case cartDirUp:
				r = '^'
			case cartDirRight:
				r = '>'
			case cartDirDown:
				r = 'v'
			case cartDirLeft:
				r = '<'
			}
			a |= ansi.RGB(64, 192, 64).FG()

		case t&cartTrack != 0:
			switch world.d[id] & cartTrackDirMask {
			case cartTrackV:
				r = '|'
				a |= ansi.RGB(96, 96, 96).FG()
			case cartTrackH:
				r = '-'
				a |= ansi.RGB(96, 96, 96).FG()
			case cartTrackX:
				r = '+'
				a |= ansi.RGB(128, 128, 128).FG()
			case cartTrackB:
				r = '\\'
				a |= ansi.RGB(96, 96, 96).FG()
			case cartTrackF:
				r = '/'
				a |= ansi.RGB(96, 96, 96).FG()
			}

		default:
			r = '?'
			a |= ansi.RGB(192, 192, 64).FG()

		}

		g.Rune[gi] = r
		g.Attr[gi] = a
	}
}

func (world *cartWorld) load(r io.Reader) error {
	if len(world.t) > 0 {
		panic("reload of world not supported")
	}

	// zero entity is zero
	world.p = append(world.p, image.ZP)
	world.t = append(world.t, 0)
	world.d = append(world.d, 0)
	world.s = append(world.s, 0)

	// scan entities from input
	sc := bufio.NewScanner(r)
	var p image.Point
	for sc.Scan() {
		line := sc.Text()
		p.X = 0
		for i := 0; i < len(line); i++ {
			var (
				t cartType
				d cartDirection
			)

			switch line[i] {
			case '|':
				t = cartTrack
				d = cartTrackV
			case '-':
				t = cartTrack
				d = cartTrackH
			case '+':
				t = cartTrack
				d = cartTrackX
			case '\\':
				t = cartTrack
				d = cartTrackB
			case '/':
				t = cartTrack
				d = cartTrackF
			case '^':
				t = cartTrack | cart
				d = cartDirUp | cartTrackV
			case '>':
				t = cartTrack | cart
				d = cartDirRight | cartTrackH
			case 'v':
				t = cartTrack | cart
				d = cartDirDown | cartTrackV
			case '<':
				t = cartTrack | cart
				d = cartDirLeft | cartTrackH
			}

			if t != 0 {
				id := len(world.t)
				world.p = append(world.p, p)
				world.t = append(world.t, t)
				world.d = append(world.d, d)
				world.s = append(world.s, 0)
				world.Index.Update(id, p)
				if t&cart != 0 {
					world.carts = append(world.carts, id)
				}
			}

			p.X++
		}

		p.Y++
	}

	// compute bounds
	world.b = image.ZR
	if len(world.p) > 1 {
		world.b.Min = world.p[1]
		world.b.Max = world.p[1]
		for _, p := range world.p[1:] {
			if world.b.Min.X > p.X {
				world.b.Min.X = p.X
			}
			if world.b.Min.Y > p.Y {
				world.b.Min.Y = p.Y
			}
			if world.b.Max.X < p.X {
				world.b.Max.X = p.X
			}
			if world.b.Max.Y < p.Y {
				world.b.Max.Y = p.Y
			}
		}
	}

	return sc.Err()
}

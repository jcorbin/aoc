package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"path"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/aoc/internal/quadindex"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		// for _, s := range []string{
		// } {
		// 	s = strings.Replace(s, "_", " ", -1)
		// 	io.WriteString(out, s)
		// }
		fmt.Fprintf(out, "\n\nUsage %s [options] [<inputFile>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

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
)

func run(in, out *os.File) error {
	var game gameUI

	if err := func() error {
		if !anansi.IsTerminal(in) {
			return game.world.load(in)
		}

		name := flag.Arg(0)
		if name == "" {
			return game.world.load(bytes.NewReader([]byte(defaultWorldData)))
		}

		f, err := os.Open(name)
		if err == nil {
			err = game.world.load(f)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		return err
	}(); err != nil {
		return err
	}

	worldMid := game.world.bounds.Size().Div(2)
	screenMid := screen.Bounds().Size().Div(2)
	game.focus = worldMid.Sub(screenMid)

	if !anansi.IsTerminal(in) {
		f, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	term := anansi.NewTerm(in, out,
		&halt,
		&resize,
		&inputReady,
		&screen,
	)
	term.SetRaw(true)
	term.AddMode(
		ansi.ModeAlternateScreen,

		ansi.ModeMouseSgrExt,
		ansi.ModeMouseBtnEvent,
		// ansi.ModeMouseAnyEvent,
	)
	resize.Send("initialize screen size")
	return term.RunWith(&game)
}

var defaultWorldData = ""

type gameWorld struct {
	bounds image.Rectangle
	quadindex.Index
	// TODO world data
}

func (world *gameWorld) describe(id int, buf *bytes.Buffer) {
	fmt.Fprintf(buf, "id:%v FIXME\n",
		id,
		// TODO world.d[id],
	)
}

func (world *gameWorld) render(g anansi.Grid, viewOffset image.Point) {
	// for id, t := range world.t {
	// 	if id == 0 { continue }

	// 	p := world.p[id]
	// 	sp := p.Add(viewOffset).Add(image.Pt(1, 1))
	// 	if sp.X < 0 || sp.Y < 0 {
	// 		continue
	// 	}

	// 	gi, ok := g.CellOffset(ansi.PtFromImage(sp))
	// 	if !ok {
	// 		continue
	// 	}

	// 	var r rune
	// 	var a ansi.SGRAttr

	// 	switch {
	// 	TODO determine from te
	// 	}

	// 	g.Rune[gi] = r
	// 	g.Attr[gi] = a
	// }
}

func (world *gameWorld) load(r io.Reader) error {

	// TODO

	// if len(world.t) > 0 {
	// 	panic("reload of world not supported")
	// }

	if nom, ok := r.(interface{ Name() string }); ok {
		log.Printf("read input from %s", nom.Name())
	}

	// scan entities from input
	sc := bufio.NewScanner(r)
	var p image.Point
	for sc.Scan() {
		line := sc.Text()
		p.X = 0
		for i := 0; i < len(line); i++ {

			switch line[i] {
			// TODO cell -> entities
			}

			p.X++
		}

		p.Y++
	}
	// compute bounds
	world.bounds = image.ZR
	// TODO
	// if len(world.p) > 1 {
	// 	world.b.Min = world.p[1]
	// 	world.b.Max = world.p[1]
	// 	for _, p := range world.p[1:] {
	// 		if world.b.Min.X > p.X {
	// 			world.b.Min.X = p.X
	// 		}
	// 		if world.b.Min.Y > p.Y {
	// 			world.b.Min.Y = p.Y
	// 		}
	// 		if world.b.Max.X < p.X {
	// 			world.b.Max.X = p.X
	// 		}
	// 		if world.b.Max.Y < p.Y {
	// 			world.b.Max.Y = p.Y
	// 		}
	// 	}
	// }

	return sc.Err()
}

func (world *gameWorld) done() bool {
	// TODO when?
	return false
}

func (world *gameWorld) tick() bool {
	if world.done() {
		return false
	}

	// TODO the thing

	return true
}

func (world *gameWorld) handleInput(term *anansi.Term, game *gameUI) error {
	// switch e {
	// TODO
	// }
	return nil
}

type gameUI struct {
	banner []byte

	timer    *time.Timer
	last     time.Time
	ticking  bool
	playing  bool
	playRate int // tick-per-second

	focus      image.Point
	viewOffset image.Point

	mess     []byte
	messSize image.Point

	world gameWorld
}

func (game *gameUI) Run(term *anansi.Term) error {
	game.timer = time.NewTimer(100 * time.Second)
	game.stopTimer()
	return term.Loop(game)
}

func (game *gameUI) setTimer(d time.Duration) {
	if game.timer == nil {
		game.timer = time.NewTimer(d)
	} else {
		game.timer.Reset(d)
	}
}

func (game *gameUI) stopTimer() {
	game.timer.Stop()
	select {
	case <-game.timer.C:
	default:
	}
}

func (game *gameUI) Update(term *anansi.Term) (redraw bool, _ error) {
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
		herr := game.handleInput(term)
		if err == nil {
			err = herr
		}
		if err != nil {
			return false, err
		}

	case now := <-game.timer.C:
		game.advance(now)
		redraw = true
	}
	return redraw, nil
}

func (game *gameUI) advance(now time.Time) {
	// no updates while displaying a message
	if !game.ticking {
		game.last = now
		return
	}

	// single-step
	if !game.playing {
		game.world.tick()
		game.last = now
		return
	}

	// advance playback
	if ticks := int(math.Round(float64(now.Sub(game.last)) / float64(time.Second) * float64(game.playRate))); ticks > 0 {
		const maxTicks = 100000
		if ticks > maxTicks {
			ticks = maxTicks
		}
		for i := 0; i < ticks; i++ {
			if !game.world.tick() {
				game.playing = false
				break
			}
		}
		game.last = now
	}

	game.ticking = true
	game.setTimer(10 * time.Millisecond) // TODO compute next time when ticks > 0; avoid spurious wakeup
}

func (game *gameUI) setBanner(mess string, args ...interface{}) {
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	game.banner = []byte(mess)
}

func (game *gameUI) handleInput(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		for _, handlerFn := range []func(e ansi.Escape, a []byte) (bool, error){
			game.handleLowInput,
			game.handleMessInput,
			game.handleWorldInput,
		} {
			if handled, err := handlerFn(e, a); err != nil {
				return err
			} else if handled {
				break
			}
		}
	}
	return nil
}

func (game *gameUI) handleLowInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	case 0x03: // stop on Ctrl-C
		return true, fmt.Errorf("read %v", e)

	case 0x0c: // clear screen on Ctrl-L
		screen.Clear()           // clear virtual contents
		screen.To(ansi.Pt(1, 1)) // cursor back to top
		screen.Invalidate()      // force full redraw
		game.setTimer(5 * time.Millisecond)
		return true, nil

	}
	return false, nil
}

func (game *gameUI) handleMessInput(e ansi.Escape, a []byte) (bool, error) {
	// no message, ignore
	if game.mess == nil {
		return false, nil
	}

	switch e {

	// <Esc> to dismiss message
	case ansi.Escape('\x1b'):
		game.setMess(nil)
		return true, nil

	// eat any other input when a message is shown
	default:
		return true, nil
	}
}

func (game *gameUI) handleWorldInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	// TODO
	// // display help
	// case ansi.Escape('?'):
	// 	return true, nil

	// arrow keys to move view
	case ansi.CUB, ansi.CUF, ansi.CUU, ansi.CUD:
		if d, ok := ansi.DecodeCursorCardinal(e, a); ok {
			p := game.focus.Add(d)
			if p.X < game.world.bounds.Min.X {
				p.X = game.world.bounds.Min.X
			}
			if p.Y < game.world.bounds.Min.Y {
				p.Y = game.world.bounds.Min.Y
			}
			if p.X >= game.world.bounds.Max.X {
				p.X = game.world.bounds.Max.X - 1
			}
			if p.Y >= game.world.bounds.Max.Y {
				p.Y = game.world.bounds.Max.Y - 1
			}
			if game.focus != p {
				game.focus = p
				game.setTimer(5 * time.Millisecond)
			}
		}
		return true, nil

	// mouse inspection
	case ansi.CSI('m'), ansi.CSI('M'):
		if m, sp, err := ansi.DecodeXtermExtendedMouse(e, a); err == nil {
			if m.ButtonID() == 1 && m.IsRelease() {
				var buf bytes.Buffer
				buf.Grow(1024)

				p := sp.ToImage().Sub(game.viewOffset)
				fmt.Fprintf(&buf, "Query @%v\n", p)

				n := 0
				cur := game.world.Index.At(p)
				for i := 0; cur.Next(); i++ {
					game.world.describe(cur.I(), &buf)
					n++
				}
				if n == 0 {
					fmt.Fprintf(&buf, "No Results\n")
				}
				fmt.Fprintf(&buf, "( <Esc> to close )")

				game.setMess(buf.Bytes())
			}
		}
		return true, nil

	// step
	case ansi.Escape('.'):
		game.setTimer(5 * time.Millisecond)
		game.ticking = true
		return true, nil

	// play/pause
	case ansi.Escape(' '):
		game.playing = !game.playing
		if !game.playing {
			game.stopTimer()
			log.Printf("pause")
		} else {
			game.last = time.Now()
			if game.playRate == 0 {
				game.playRate = 1
			}
			game.ticking = true
			log.Printf("play at %v ticks/s", game.playRate)
		}
		game.setTimer(5 * time.Millisecond)
		return true, nil

	// speed control
	case ansi.Escape('+'):
		game.playRate *= 2
		log.Printf("speed up to %v ticks/s", game.playRate)
		game.setTimer(5 * time.Millisecond)
		return true, nil
	case ansi.Escape('-'):
		rate := game.playRate / 2
		if rate <= 0 {
			rate = 1
		}
		if game.playRate != rate {
			game.playRate = rate
			log.Printf("slow down to %v ticks/s", game.playRate)
		}
		game.setTimer(5 * time.Millisecond)
		return true, nil

	}

	return false, nil
}

func (game *gameUI) WriteTo(w io.Writer) (n int64, err error) {
	screen.Clear()
	game.world.render(screen.Grid, game.viewOffset)
	overlayLogs()
	game.overlayBanner()
	game.overlayMess()
	return screen.WriteTo(w)
}

func (game *gameUI) setMess(mess []byte) {
	game.mess = mess
	if mess == nil {
		game.messSize = image.ZP
	} else {
		game.messSize = measureTextBox(mess).Size()
	}
	game.setTimer(5 * time.Millisecond)
}

func (game *gameUI) overlayBanner() {
	at := screen.Grid.Rect.Min
	bannerWidth := measureTextBox(game.banner).Dx()
	screenWidth := screen.Bounds().Dx()
	at.X += screenWidth/2 - bannerWidth/2
	writeIntoGrid(screen.Grid.SubAt(at), game.banner)
}

func (game *gameUI) overlayMess() {
	if game.mess == nil || game.messSize == image.ZP {
		return
	}
	screenSize := screen.Bounds().Size()
	screenMid := screenSize.Div(2)
	messMid := game.messSize.Div(2)
	offset := screenMid.Sub(messMid)
	writeIntoGrid(screen.Grid.SubAt(screen.Grid.Rect.Min.Add(offset)), game.mess)
}

func writeIntoGrid(g anansi.Grid, b []byte) {
	var cur anansi.CursorState
	cur.Point = g.Rect.Min
	for len(b) > 0 {
		e, a, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e == 0 {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			e = ansi.Escape(r)
		}
		switch e {
		case ansi.Escape('\n'):
			cur.Y++
			cur.X = g.Rect.Min.X

		case ansi.CSI('m'):
			if attr, _, err := ansi.DecodeSGR(a); err == nil {
				cur.MergeSGR(attr)
			}

		default:
			// write runes into grid, with cursor style, ignoring any other
			// escapes; treating `_` as transparent
			if !e.IsEscape() {
				if i, ok := g.CellOffset(cur.Point); ok {
					if e != ansi.Escape('_') {
						g.Rune[i] = rune(e)
						g.Attr[i] = cur.Attr
					}
				}
				cur.X++
			}
		}
	}
}

func measureTextBox(b []byte) (box ansi.Rectangle) {
	box.Min = ansi.Pt(1, 1)
	box.Max = ansi.Pt(1, 1)
	pt := box.Min
	for len(b) > 0 {
		e, _ /*a*/, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e == 0 {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			e = ansi.Escape(r)
		}
		switch e {

		case ansi.Escape('\n'):
			pt.Y++
			pt.X = 1

			// TODO would be nice to borrow cursor movement processing from anansi.Screen et al

		default:
			// ignore escapes, advance on runes
			if !e.IsEscape() {
				pt.X++
			}

		}
		if box.Max.X < pt.X {
			box.Max.X = pt.X
		}
		if box.Max.Y < pt.Y {
			box.Max.Y = pt.Y
		}
	}
	return box
}

// in-memory log buffer
var logs bytes.Buffer

func init() {
	f, err := os.Create("game.log")
	if err != nil {
		log.Fatalf("failed to create game.log: %v", err)
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

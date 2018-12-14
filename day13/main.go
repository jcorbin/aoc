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
	"sort"
	"strings"
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
		for _, s := range []string{
			welcomeMess,
			usageMess,
			inputMess,
			helpMessFooter,
		} {
			s = strings.Replace(s, "_", " ", -1)
			io.WriteString(out, s)
		}
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

var (
	cartBTurns = [4]cartDirection{cartDirLeft, cartDirDown, cartDirRight, cartDirUp}
	cartFTurns = [4]cartDirection{cartDirRight, cartDirUp, cartDirLeft, cartDirDown}
)

const (
	cartDirMask      cartDirection = 0x01 | 0x02
	cartTrackDirMask cartDirection = 0x04 | 0x08 | 0x10 | 0x20
)

type cartWorld struct {
	quadindex.Index
	b image.Rectangle // world bounds
	p []image.Point   // location
	t []cartType      // is cart? has track?
	d []cartDirection // direction of cart and/or track here
	s []int           // cart state

	// mode for part 2
	lastStanding bool

	carts   []int
	crashed bool

	last     time.Time
	ticking  bool
	playing  bool
	playRate int // tick-per-second

	timer *time.Timer

	viewOffset image.Point

	debug  bool
	debugP image.Point

	mess     []byte
	messSize image.Point
}

var traceFlag = flag.Bool("trace", false, "log trace events")

func run(in, out *os.File) error {
	var world cartWorld
	world.lastStanding = true // TODO improve this ux

	helpMess += welcomeMess + keysMess

	if err := func() error {
		if !anansi.IsTerminal(in) {
			return world.load(in)
		}

		name := flag.Arg(0)
		if name == "" {
			helpMess += inputMessEx + inputMess
			return world.load(bytes.NewReader([]byte(exProblem)))
		}

		f, err := os.Open(name)
		if err == nil {
			err = world.load(f)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		return err
	}(); err != nil {
		return err
	}

	helpMess += helpMessFooter

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
	return term.RunWith(&world)
}

var welcomeMess = "" +
	`________________/---------\` + "\n" +
	`/---------------/ Welcome \--------------\` + "\n" +
	`|              To Cart World!            |` + "\n" +
	`+----------------------------------------+` + "\n" +
	`| A simulation done for AoC 2018 Day 13  |` + "\n" +
	`|  https://adventofcode.com/2018/day/13  |` + "\n" +
	`|  https://github.com/jcorbin/aoc        |` + "\n"

var keysMess = "" +
	`+----------------------------------------+` + "\n" +
	`| Keys:                                  |` + "\n" +
	`|   <Esc>   to dismiss this help message |` + "\n" +
	`|   ?       to display it again          |` + "\n" +
	`|   .       to single step the world     |` + "\n" +
	`|   <Space> to play/pause the simulation |` + "\n" +
	`|   +/-     to control play speed        |` + "\n" +
	`|                                        |` + "\n" +
	`| Click mouse to inspect cell            |` + "\n"

var usageMess = "" +
	`+----------------------------------------+` + "\n" +
	`| If no input is given, the trivial      |` + "\n" +
	`| 3-square 2-cart example is used.       |` + "\n"

var inputMessEx = "" +
	`+----------------------------------------+` + "\n" +
	`| Using canned example input.            |` + "\n"

var inputMess = "" +
	`| Problem input may be given on stdin,   |` + "\n" +
	`| Or by passing a file argument.         |` + "\n"

var helpMessFooter = "" +
	`\----------------------------------------/`

var exProblem = "" +
	`/->-\` + "\n" +
	`|   |  /----\` + "\n" +
	`| /-+--+-\  |` + "\n" +
	`| | |  | v  |` + "\n" +
	`\-+-/  \-+--/` + "\n" +
	`  \------/`

var helpMess string

func (world *cartWorld) Run(term *anansi.Term) error {
	world.timer = time.NewTimer(100 * time.Second)
	world.setMess([]byte(helpMess))
	world.stopTimer()
	return term.Loop(world)
}

func (world *cartWorld) setTimer(d time.Duration) {
	if world.timer == nil {
		world.timer = time.NewTimer(d)
	} else {
		world.timer.Reset(d)
	}
}

func (world *cartWorld) stopTimer() {
	world.timer.Stop()
	select {
	case <-world.timer.C:
	default:
	}
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
		herr := world.handleInput(term)
		if err == nil {
			err = herr
		}
		if err != nil {
			return false, err
		}

	case now := <-world.timer.C:
		world.update(now)
		redraw = true
	}
	return redraw, nil
}

func (world *cartWorld) update(now time.Time) {
	// no updates while displaying a message
	if !world.ticking {
		world.last = now
		return
	}

	// single-step
	if !world.playing {
		world.tick()
		world.last = now
		return
	}

	const maxTicks = 100000
	elapsed := now.Sub(world.last)
	ticks := int(math.Round(float64(elapsed) / float64(time.Second) * float64(world.playRate)))
	if ticks > maxTicks {
		ticks = maxTicks
	}

	for i := 0; i < ticks; i++ {
		if !world.tick() {
			world.playing = false
			break
		}
	}

	world.ticking = true
	world.setTimer(10 * time.Millisecond) // TODO compute next time when ticks > 0; avoid spurious wakeup
}

func (world *cartWorld) done() bool {
	if world.debug {
		return true
	}
	if world.lastStanding {
		switch len(world.carts) {
		case 0:
			log.Printf("NOTHING LEFT?!?")
			return true
		case 1:
			id := world.carts[0]
			log.Printf("last @%v", world.p[id])
			return true

		default:
			return false
		}
	}
	return world.crashed || len(world.carts) == 0
}

func (world *cartWorld) tick() bool {
	if world.done() {
		return false
	}
	world.pruneCarts()

	anyRemoved := false
	for carti, id := range world.carts {
		t := world.t[id]
		if t&cart == 0 {
			continue
		}

		p := world.p[id]
		d := world.d[id] & cartDirMask
		// log.Printf("consider id:%v@%v d:%v", id, p, d)

		var dest image.Point
		switch d {
		case cartDirUp:
			dest = p.Add(image.Pt(0, -1))
		case cartDirRight:
			dest = p.Add(image.Pt(1, 0))
		case cartDirDown:
			dest = p.Add(image.Pt(0, 1))
		case cartDirLeft:
			dest = p.Add(image.Pt(-1, 0))
		default:
			log.Printf("BOGUS DIR: id:%v d:%v", id, d)
			world.setDebug(p)
			return false
		}

		cur := world.Index.At(dest)
		cur.Next()
		tid := cur.I()
		if tid < 0 {
			log.Printf("NOWHERE id:%v@%v to:%v", id, p, dest)
			world.setDebug(p)
			return false
		}

		destT := world.t[tid]

		if destT&cart != 0 {
			world.removeCart(id)
			world.removeCart(tid)
			if world.lastStanding {
				log.Printf("removed @%v", dest)
				anyRemoved = true
			} else {
				world.t[tid] |= cartCrash
				log.Printf("CRASH @%v", dest)
				world.crashed = true
			}
			continue
		}

		if destT&cartTrack == 0 {
			world.removeCart(id)
			log.Printf("LIMBO @%v", dest)
			world.setDebug(dest)
			return false
		}

		s := world.s[id]

		var trec struct {
			id, tid int
			op, np  image.Point
			od, nd  cartDirection
			os, ns  int
		}
		if *traceFlag {
			trec.id = id
			trec.tid = tid
			trec.op = p
			trec.np = dest
			trec.od = d
			trec.os = s
		}

		world.removeCart(id)

		td := world.d[tid] & cartTrackDirMask
		switch td {

		case cartTrackX: // intersections
			switch s % 3 {
			case 0: // left
				d = (d - 1 + 4) % 4
			case 1: // straight
			case 2: // right
				d = (d + 1) % 4
			}
			s++

		case cartTrackB: // `\` corner
			d = cartBTurns[d]

		case cartTrackF: // `/` corner
			d = cartFTurns[d]

		}

		if *traceFlag {
			trec.nd, trec.ns = d, s
			log.Printf("+ %v", trec)
		}

		world.t[tid] |= cart
		world.d[tid] |= d
		world.s[tid] = s
		world.carts[carti] = tid
	}

	world.pruneCarts()

	if anyRemoved {
		log.Printf("remaining: %v", len(world.carts))
	}

	return true
}

func (world *cartWorld) setDebug(p image.Point) {
	world.debug = true
	world.debugP = p
	if world.playing {
		world.playing = false
	}
}

func (world *cartWorld) pruneCarts() {
	i := 0
	for j := 0; j < len(world.carts); j++ {
		id := world.carts[j]
		if id == 0 {
			continue
		}
		t := world.t[id]
		if t&cart == 0 {
			world.carts[j] = 0
			continue
		}
		world.carts[i] = world.carts[j]
		i++
	}
	world.carts = world.carts[:i]

	sort.Slice(world.carts, func(i, j int) bool {
		pi := world.p[world.carts[i]]
		pj := world.p[world.carts[j]]
		if pi.Y < pj.Y {
			return true
		}
		if pi.Y > pj.Y {
			return false
		}
		return pi.X < pj.X
	})
}

func (world *cartWorld) removeCart(id int) {
	world.t[id] &= ^cart
	world.d[id] &= ^cartDirMask
	world.s[id] = 0
}

func (world *cartWorld) handleInput(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		for _, handlerFn := range []func(e ansi.Escape, a []byte) (bool, error){
			world.handleLowInput,
			world.handleMessInput,
			world.handleSimInput,
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

func (world *cartWorld) handleLowInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	case 0x03: // stop on Ctrl-C
		return true, fmt.Errorf("read %v", e)

	case 0x0c: // clear screen on Ctrl-L
		screen.Clear()           // clear virtual contents
		screen.To(ansi.Pt(1, 1)) // cursor back to top
		screen.Invalidate()      // force full redraw
		return true, nil

	}
	return false, nil
}

func (world *cartWorld) handleMessInput(e ansi.Escape, a []byte) (bool, error) {
	// no message, ignore
	if world.mess == nil {
		return false, nil
	}

	switch e {

	// <Esc> to dismiss message
	case ansi.Escape('\x1b'):
		world.setMess(nil)
		return true, nil

	// eat any other input when a message is shown
	default:
		return true, nil
	}
}

func (world *cartWorld) handleSimInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {
	// display help
	case ansi.Escape('?'):
		world.setMess([]byte(helpMess))
		return true, nil

	// mouse inspection
	case ansi.CSI('m'), ansi.CSI('M'):
		if m, sp, err := ansi.DecodeXtermExtendedMouse(e, a); err == nil {
			if m.ButtonID() == 1 && m.IsRelease() {
				var buf bytes.Buffer
				buf.Grow(1024)

				p := sp.ToImage().Sub(world.viewOffset)
				fmt.Fprintf(&buf, "Query @%v\n", p)

				n := 0
				cur := world.Index.At(p)
				for i := 0; cur.Next(); i++ {
					id := cur.I()
					fmt.Fprintf(&buf, "id:%v p:%v t:%v d:%v s:%v\n",
						id,
						world.p[id],
						world.t[id],
						world.d[id],
						world.s[id],
					)
					n++
				}
				if n == 0 {
					fmt.Fprintf(&buf, "No Results\n")
				}
				fmt.Fprintf(&buf, "( <Esc> to close )")

				world.setMess(buf.Bytes())
			}
		}
		return true, nil

	// step
	case ansi.Escape('.'):
		world.setTimer(5 * time.Millisecond)
		world.ticking = true
		return true, nil

	// play/pause
	case ansi.Escape(' '):
		world.playing = !world.playing
		if !world.playing {
			world.stopTimer()
			log.Printf("pause")
		} else {
			world.last = time.Now()
			if world.playRate == 0 {
				world.playRate = 1
			}
			world.setTimer(10 * time.Millisecond)
			world.ticking = true
			log.Printf("play at %v ticks/s", world.playRate)
		}
		return true, nil

	// speed control
	case ansi.Escape('+'):
		if world.playing {
			world.playRate *= 2
			log.Printf("speed up to %v ticks/s", world.playRate)
		}
		return true, nil
	case ansi.Escape('-'):
		if world.playing {
			rate := world.playRate / 2
			if rate <= 0 {
				rate = 1
			}
			if world.playRate != rate {
				world.playRate = rate
				log.Printf("slow down to %v ticks/s", world.playRate)
			}
		}
		return true, nil

	}
	return false, nil
}

func (world *cartWorld) WriteTo(w io.Writer) (n int64, err error) {
	screen.Clear()
	world.render(screen.Grid)
	overlayLogs()
	world.overlayMess()
	return screen.WriteTo(w)
}

func (world *cartWorld) setMess(mess []byte) {
	world.mess = mess
	if mess == nil {
		world.messSize = image.ZP
	} else {
		world.messSize = measureTextBox(mess).Size()
	}
	world.setTimer(5 * time.Millisecond)
}

func (world *cartWorld) overlayMess() {
	if world.mess == nil || world.messSize == image.ZP {
		return
	}
	screenSize := screen.Bounds().Size()
	screenMid := screenSize.Div(2)
	messMid := world.messSize.Div(2)

	offset := screenMid.Sub(messMid)
	b := world.mess
	var cur anansi.CursorState
	cur.Point = ansi.Pt(1, 1).Add(offset)
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
			cur.X = 1 + offset.X

		case ansi.CSI('m'):
			if attr, _, err := ansi.DecodeSGR(a); err == nil {
				cur.MergeSGR(attr)
			}

		default:
			// write runes into screen grid, with cursor style, ignoring any
			// other escapes; treating `_` as transparent
			if !e.IsEscape() {
				if i, ok := screen.Grid.CellOffset(cur.Point); ok {
					if e != ansi.Escape('_') {
						screen.Grid.Rune[i] = rune(e)
						screen.Grid.Attr[i] = cur.Attr
					}
				}
				cur.X++
			}
		}
	}
}

func (world *cartWorld) render(g anansi.Grid) {
	var (
		debugColor  = ansi.RGB(96, 32, 16)
		crashColor  = ansi.RGB(192, 64, 64)
		cartColor   = ansi.RGB(64, 192, 64)
		trackColor  = ansi.RGB(64, 64, 64)
		trackXColor = ansi.RGB(128, 128, 128)
		unkColor    = ansi.RGB(192, 192, 64)
	)

	if world.debug {
		bnd := g.Bounds()
		m := bnd.Size().Div(2)
		world.viewOffset = m.Sub(world.debugP)
	}

	for id, t := range world.t {
		if id == 0 {
			continue
		}

		p := world.p[id]
		sp := p.Add(world.viewOffset).Add(image.Pt(1, 1))
		if sp.X < 1 || sp.Y < 1 {
			continue
		}

		gi, ok := g.CellOffset(ansi.PtFromImage(sp))
		if !ok {
			continue
		}

		var r rune
		var a ansi.SGRAttr

		if world.debug && p == world.debugP {
			a |= debugColor.BG()
		}

		switch {

		case t&cartCrash != 0:
			r = 'X'
			a |= crashColor.FG()

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
			a |= cartColor.FG()

		case t&cartTrack != 0:
			switch world.d[id] & cartTrackDirMask {
			case cartTrackV:
				r = '|'
				a |= trackColor.FG()
			case cartTrackH:
				r = '-'
				a |= trackColor.FG()
			case cartTrackX:
				r = '+'
				a |= trackXColor.FG()
			case cartTrackB:
				r = '\\'
				a |= trackColor.FG()
			case cartTrackF:
				r = '/'
				a |= trackColor.FG()
			}

		default:
			r = '?'
			a |= unkColor.FG()

		}

		g.Rune[gi] = r
		g.Attr[gi] = a
	}
}

func (world *cartWorld) load(r io.Reader) error {
	if len(world.t) > 0 {
		panic("reload of world not supported")
	}

	if nom, ok := r.(interface{ Name() string }); ok {
		log.Printf("read input from %s", nom.Name())
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

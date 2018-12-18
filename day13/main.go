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
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/aoc/internal/layerui"
	"github.com/jcorbin/aoc/internal/quadindex"
)

var logfile = flag.String("logfile", "", "log file")

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

	log.SetFlags(log.Ltime | log.Lmicroseconds)
	layerui.MustOpenLogFile(*logfile)

	anansi.MustRun(func() error {
		var world cartWorld

		if err := world.Init(); err != nil {
			return err
		}

		in, out := layerui.MustOpenTermFiles(os.Stdin, os.Stdout)

		world.WorldLayer.World = &world

		ui := layerui.LayerUI{
			Layers: []layerui.Layer{
				&layerui.LogLayer{SubGrid: func(g anansi.Grid, numLines int) anansi.Grid {
					if numLines > 5 {
						numLines = 5
					}
					return g.SubAt(ansi.Pt(
						1, g.Bounds().Dy()-numLines,
					))
				}},
				&world.ModalLayer,
				&world.BannerLayer,
				&world.WorldLayer,
			},
		}
		ui.SetupSignals()
		term := anansi.NewTerm(in, out, &ui)
		term.SetRaw(true)
		term.AddMode(
			ansi.ModeAlternateScreen,
			ansi.ModeMouseSgrExt,
			ansi.ModeMouseBtnEvent,
			// ansi.ModeMouseAnyEvent, TODO option
		)

		return term.RunWithFunc(func(term *anansi.Term) error {
			return term.Loop(&ui)
		})
	}())
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
	layerui.ModalLayer
	layerui.BannerLayer
	layerui.WorldLayer

	quadindex.Index
	b image.Rectangle // world bounds
	p []image.Point   // location
	t []cartType      // is cart? has track?
	d []cartDirection // direction of cart and/or track here
	s []int           // cart state

	carts []int

	needsDraw  time.Duration
	crash      int
	autoRemove bool
	hi         bool
	hiStop     bool
	hiAt       image.Point
}

var traceFlag = flag.Bool("trace", false, "log trace events")

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
	`|   X       to clear a crash             |` + "\n" +
	`|   *       toggle auto remove           |` + "\n" +
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

func (world *cartWorld) Init() (err error) {
	helpMess += welcomeMess + keysMess
	if !anansi.IsTerminal(os.Stdin) {
		log.Printf("load stdin")
		err = world.load(os.Stdin)
	} else if name := flag.Arg(0); name != "" {
		log.Printf("load file arg %q", name)
		f, err := os.Open(name)
		if err == nil {
			err = world.load(f)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
	} else {
		log.Printf("load builtin")
		helpMess += inputMessEx + inputMess
		err = world.load(bytes.NewReader([]byte(exProblem)))
	}
	helpMess += helpMessFooter

	world.SetFocus(world.b.Size().Div(2))
	world.Display(helpMess)
	return err
}

func (world *cartWorld) Bounds() image.Rectangle { return world.b }

func (world *cartWorld) done() bool {
	if world.hiStop {
		return true
	}
	switch len(world.carts) {
	case 0:
		world.Say("No Carts Left")
		return true
	case 1:
		p := world.p[world.carts[0]]
		world.setHighlight(true, p, "last @%v", p)
		return true
	}
	return false
}

func (world *cartWorld) Tick() bool {
	if world.done() {
		return false
	}
	world.pruneCarts()

	var removed []image.Point
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
			world.setHighlight(true, p, "INVALID direction: id:%v d:%v", id, d)
			return false
		}

		cur := world.Index.At(dest)
		cur.Next()
		tid := cur.I()
		if tid < 0 {
			world.setHighlight(true, p, "INVALID move id:%v@%v to:%v", id, p, dest)
			return false
		}

		destT := world.t[tid]

		if destT&cart != 0 {
			world.removeCart(id)
			world.removeCart(tid)
			if world.autoRemove {
				removed = append(removed, dest)
			} else {
				world.t[tid] |= cartCrash
				world.crash = tid
				world.setHighlight(true, dest, "CRASH @%v ( press X to clear )", dest)
			}
			continue
		}

		if destT&cartTrack == 0 {
			world.removeCart(id)
			world.setHighlight(true, p, "LIMBO cart @%v", dest)
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

	if len(removed) > 0 {
		log.Printf("removed (auto) @%v, remaining: %v", removed, len(world.carts))
	}

	return true
}

func (world *cartWorld) setHighlight(stop bool, at image.Point, mess string, args ...interface{}) {
	world.Say(fmt.Sprintf(mess, args...))
	world.hi = true
	world.hiStop = stop
	world.hiAt = at
	world.Pause()
	world.SetFocus(world.hiAt)
}

func (world *cartWorld) clearHighlight() {
	world.Say("")
	world.hi = false
	world.hiStop = false
	world.hiAt = image.ZP
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

// NeedsDraw returns non-zero if the world needs to be drawn.
func (world *cartWorld) NeedsDraw() time.Duration {
	return world.needsDraw
}

func (world *cartWorld) HandleInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {
	// display help
	case ansi.Escape('?'):
		world.Display(helpMess)
		return true, nil

	// mouse inspection
	case ansi.CSI('m'), ansi.CSI('M'):
		if m, sp, err := ansi.DecodeXtermExtendedMouse(e, a); err == nil {
			if m.ButtonID() == 1 && m.IsRelease() {
				var buf bytes.Buffer
				buf.Grow(1024)

				p := sp.ToImage().Sub(world.ViewOffset())
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

				world.Display(buf.String())
			}
		}
		return true, nil

	// clear crash
	case ansi.Escape('X'):
		if world.crash != 0 {
			world.t[world.crash] &= ^cartCrash
			log.Printf("removed @%v, remaining: %v", world.p[world.crash], len(world.carts))
			world.crash = 0
			world.Pause()
			world.clearHighlight()
			world.needsDraw = 5 * time.Millisecond
		}
		return true, nil

	// toggle auto remove
	case ansi.Escape('*'):
		world.autoRemove = !world.autoRemove
		if world.autoRemove {
			log.Printf("auto remove on")
		} else {
			log.Printf("auto remove on")
		}
		world.needsDraw = 5 * time.Millisecond
		return true, nil

	}
	return false, nil
}

func (world *cartWorld) Render(g anansi.Grid, viewOffset image.Point) {
	var (
		hiColor     = ansi.RGB(96, 32, 16)
		crashColor  = ansi.RGB(192, 64, 64)
		cartColor   = ansi.RGB(64, 192, 64)
		trackColor  = ansi.RGB(64, 64, 64)
		trackXColor = ansi.RGB(128, 128, 128)
		unkColor    = ansi.RGB(192, 192, 64)
	)

	for id, t := range world.t {
		if id == 0 {
			continue
		}

		p := world.p[id]
		sp := p.Add(viewOffset).Add(image.Pt(1, 1))
		if sp.X < 0 || sp.Y < 0 {
			continue
		}

		gi, ok := g.CellOffset(ansi.PtFromImage(sp))
		if !ok {
			continue
		}

		var r rune
		var a ansi.SGRAttr

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

	if world.hi {
		sp := world.hiAt.Add(viewOffset).Add(image.Pt(1, 1))
		if sp.X >= 0 && sp.Y >= 0 {
			if gi, ok := g.CellOffset(ansi.PtFromImage(sp)); ok {
				g.Attr[gi] = hiColor.BG()
			}
		}
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

package main

import (
	"bufio"
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

	"aoc/internal/geom"
	"aoc/internal/infernio"
	"aoc/internal/layerui"
	"aoc/internal/quadindex"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
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

	anansi.MustRun(layerui.WithOpenLogFile(*logfile, run))
}

func run() error {
	var world cartWorld
	world.ui.LogLayer.SubGrid = layerui.BottomNLines(5)
	world.ui.ViewLayer.Client = &world
	world.ui.WorldLayer.View = &world.ui.ViewLayer
	world.ui.WorldLayer.World = &world

	if err := infernio.LoadInput(builtinInput, world.load); err != nil {
		return err
	}

	return layerui.Run(
		&world.ui.ModalLayer,
		&world.ui.BannerLayer,
		&world.ui.LogLayer,
		&world.ui.WorldLayer,

		// TODO for inspecting
		// ansi.ModeMouseSgrExt,
		// ansi.ModeMouseBtnEvent,
	)
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
	ui struct {
		layerui.ModalLayer
		layerui.BannerLayer
		layerui.LogLayer
		layerui.ViewLayer
		layerui.WorldLayer
	}

	quadindex.Index
	b image.Rectangle // world bounds
	p []image.Point   // location
	t []cartType      // is cart? has track?
	d []cartDirection // direction of cart and/or track here
	s []int           // cart state

	carts []int

	helpMess   string
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
	`| Using default example input.           |` + "\n"

var inputMessFrom = "" +
	`+----------------------------------------+` + "\n" +
	`| Read input from: % 21s |` + "\n"

var messLine = `| % 38s |` + "\n"

var inputMess = "" +
	`| Problem input may be given on stdin,   |` + "\n" +
	`| Or by passing a file argument.         |` + "\n"

func buildInputMess(name string) string {
	var s string
	switch name {
	case builtinInput.Name():
		s = inputMessEx

	default:
		n := len(name) // TODO runes not bytes
		if len(name) <= 21 {
			s = fmt.Sprintf(inputMessFrom, name)
		} else {
			if n > 38 {
				l, r := n/2-1, (n+1)/2
				name = fmt.Sprintf("%s…%s", name[:l], name[:r])
			}
			s = fmt.Sprintf(inputMessFrom, "")
			s += fmt.Sprintf(messLine, name)
		}
	}

	return s + inputMess
}

var helpMessFooter = "" +
	`\----------------------------------------/`

var builtinInput = infernio.Builtin("" +
	`/->-\` + "\n" +
	`|   |  /----\` + "\n" +
	`| /-+--+-\  |` + "\n" +
	`| | |  | v  |` + "\n" +
	`\-+-/  \-+--/` + "\n" +
	`  \------/`)

func (world *cartWorld) Bounds() image.Rectangle { return world.b }

func (world *cartWorld) done() bool {
	if world.hiStop {
		return true
	}
	switch len(world.carts) {
	case 0:
		world.ui.Say("No Carts Left")
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
	world.ui.Say(fmt.Sprintf(mess, args...))
	world.hi = true
	world.hiStop = stop
	world.hiAt = at
	world.ui.Pause()
	world.ui.SetFocus(world.hiAt)
}

func (world *cartWorld) clearHighlight() {
	world.ui.Say("")
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
		world.ui.Display(world.helpMess)
		return true, nil

	/* mouse inspection TODO bring back as an independent layer
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
	*/

	// toggle auto remove
	case ansi.Escape('*'):
		world.autoRemove = !world.autoRemove
		if !world.autoRemove {
			log.Printf("auto remove off")
			world.needsDraw = time.Millisecond
			return true, nil
		}
		log.Printf("auto remove on")
		fallthrough

	// clear crash
	case ansi.Escape('X'):
		if world.crash != 0 {
			world.t[world.crash] &= ^cartCrash
			log.Printf("removed @%v, remaining: %v", world.p[world.crash], len(world.carts))
			world.crash = 0
			world.ui.Pause()
			world.clearHighlight()
			world.needsDraw = time.Millisecond
		}
		return true, nil

	}
	return false, nil
}

func (world *cartWorld) Render(g anansi.Grid, viewport image.Rectangle) {
	var (
		hiColor     = ansi.RGB(96, 32, 16)
		crashColor  = ansi.RGB(192, 64, 64)
		cartColor   = ansi.RGB(64, 192, 64)
		trackColor  = ansi.RGB(64, 64, 64)
		trackXColor = ansi.RGB(128, 128, 128)
		unkColor    = ansi.RGB(192, 192, 64)
	)

	viewOffset := g.Rect.Min.ToImage().Sub(viewport.Min)

	for id, t := range world.t {
		if id == 0 {
			continue
		}

		p := world.p[id]
		if !p.In(viewport) {
			continue
		}

		gi, ok := g.CellOffset(ansi.PtFromImage(p.Add(viewOffset)))
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

	if world.hi && world.hiAt.In(viewport) {
		if gi, ok := g.CellOffset(ansi.PtFromImage(world.hiAt.Add(viewOffset))); ok {
			g.Attr[gi] = hiColor.BG()
		}
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
		world.b = geom.PointRect(world.p[1])
		for _, p := range world.p[1:] {
			world.b = world.b.Union(geom.PointRect(p))
		}
	}

	err := sc.Err()
	if err == nil {
		world.ui.SetFocus(world.b.Size().Div(2))
		world.helpMess = "" +
			welcomeMess +
			keysMess +
			buildInputMess(infernio.ReaderName(r)) +
			helpMessFooter
		world.ui.Display(world.helpMess)
	}
	return err
}

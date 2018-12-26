package main

import (
	"aoc/internal/geom"
	"aoc/internal/infernio"
	"aoc/internal/layerui"
	"bufio"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	justSolve = flag.Bool("solve", false, "just solve it")
	logfile   = flag.String("logfile", "", "log file")
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	flag.Parse()
	anansi.MustRun(layerui.WithOpenLogFile(*logfile, run))
}

var builtinInput = infernio.Builtin("" +
	"position=< 9,  1> velocity=< 0,  2>\n" +
	"position=< 7,  0> velocity=<-1,  0>\n" +
	"position=< 3, -2> velocity=<-1,  1>\n" +
	"position=< 6, 10> velocity=<-2, -1>\n" +
	"position=< 2, -4> velocity=< 2,  2>\n" +
	"position=<-6, 10> velocity=< 2, -2>\n" +
	"position=< 1,  8> velocity=< 1, -1>\n" +
	"position=< 1,  7> velocity=< 1,  0>\n" +
	"position=<-3, 11> velocity=< 1, -2>\n" +
	"position=< 7,  6> velocity=<-1, -1>\n" +
	"position=<-2,  3> velocity=< 1,  0>\n" +
	"position=<-4,  3> velocity=< 2,  0>\n" +
	"position=<10, -3> velocity=<-1,  1>\n" +
	"position=< 5, 11> velocity=< 1, -2>\n" +
	"position=< 4,  7> velocity=< 0, -1>\n" +
	"position=< 8, -2> velocity=< 0,  1>\n" +
	"position=<15,  0> velocity=<-2,  0>\n" +
	"position=< 1,  6> velocity=< 1,  0>\n" +
	"position=< 8,  9> velocity=< 0, -1>\n" +
	"position=< 3,  3> velocity=<-1,  1>\n" +
	"position=< 0,  5> velocity=< 0, -1>\n" +
	"position=<-2,  2> velocity=< 2,  0>\n" +
	"position=< 5, -2> velocity=< 1,  2>\n" +
	"position=< 1,  4> velocity=< 2,  1>\n" +
	"position=<-2,  7> velocity=< 2, -2>\n" +
	"position=< 3,  6> velocity=<-1, -1>\n" +
	"position=< 5,  0> velocity=< 1,  0>\n" +
	"position=<-6,  0> velocity=< 2,  0>\n" +
	"position=< 5,  9> velocity=< 1, -2>\n" +
	"position=<14,  7> velocity=<-2,  0>\n" +
	"position=<-3,  6> velocity=< 2, -1>\n")

const helpMess = "" +
	`+----------------------------------------+` + "\n" +
	`| Keys:                                  |` + "\n" +
	`|   <Esc>   to dismiss this help message |` + "\n" +
	`|   ?       to display it again          |` + "\n" +
	`|   .       to single step               |` + "\n" +
	`|   <Space> to play/pause                |` + "\n" +
	`|   +/-     to control play speed        |` + "\n" +
	`|   </>     to change time direction     |` + "\n" +
	`|   *       to solve (NOTE: T++ only)    |` + "\n" +
	`+----------------------------------------+`

func run() error {
	var sp space

	if err := infernio.LoadInput(builtinInput, sp.load); err != nil {
		return err
	}

	if *justSolve || !anansi.IsTerminal(os.Stdout) {
		return solve(sp, os.Stdout)
	}

	var world spaceWorld

	world.ui.LogLayer.SubGrid = layerui.BottomNLines(5)
	world.ui.WorldLayer.World = &world
	world.ui.ViewLayer.Client = world.ui.WorldLayer.World
	world.ui.WorldLayer.View = &world.ui.ViewLayer

	world.space = &sp
	world.tick = world.space.step
	world.updateBanner()
	world.ui.Display(helpMess)

	return layerui.Run(
		&world.ui.ModalLayer,
		&world.ui.BannerLayer,
		&world.ui.LogLayer,
		&world.ui.WorldLayer,
	)
}

func solve(sp space, out *os.File) error {
	lastn := 0
	for {
		sz := sp.b.Size()
		n := sz.X * sz.Y
		var dn int
		if lastn != 0 {
			dn = n - lastn
			if dn >= 0 {
				break
			}
		}
		lastn = n
		sp.step()
	}

	sp.back()
	fmt.Fprintf(out, "--- T:%v %v\r\n", sp.t, sp.b.Size())
	bi := sp.render()
	anansi.WriteBitmap(out, bi)
	fmt.Fprintf(out, "\r\n")

	return nil
}

type spaceWorld struct {
	ui struct {
		layerui.ModalLayer
		layerui.BannerLayer
		layerui.LogLayer
		layerui.ViewLayer
		layerui.WorldLayer
	}

	*space
	rendered  int
	lastT     int
	lastSize  image.Point
	tick      func()
	reverse   bool
	post      func(*spaceWorld)
	needsDraw time.Duration
}

func (world *spaceWorld) Bounds() image.Rectangle {
	sz := world.b.Size()
	sz.X = (sz.X + 1) / 2
	sz.Y = (sz.Y + 3) / 4
	return image.Rectangle{image.ZP, sz.Add(image.Pt(1, 1))}
}

func (world *spaceWorld) NeedsDraw() time.Duration { return world.needsDraw }

func (world *spaceWorld) updateBanner() {
	sz := world.b.Size()
	foc := sz.Div(2)
	foc.X = (foc.X + 1) / 2
	foc.Y = (foc.Y + 3) / 4
	dir := "++"
	if world.reverse {
		dir = "--"
	}
	mess := fmt.Sprintf("T:%v%s size:%v bounds:%v", world.t, dir, sz, world.b)
	world.ui.Say(mess)
	world.ui.SetFocus(foc)
}

func (world *spaceWorld) Tick() bool {
	world.lastT = world.t
	world.lastSize = world.b.Size()
	world.tick()
	world.updateBanner()
	world.check()
	if world.lastT != world.t {
		world.lastT = world.t
		world.lastSize = world.b.Size()
	}
	return true
}

func (world *spaceWorld) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	switch e {

	case '?':
		world.ui.Display(helpMess)
		return true, nil

	case '<':
		if !world.reverse {
			world.reverse = true
			world.tick = world.space.back
		}
		world.Tick()
		world.needsDraw = time.Millisecond
		return true, nil

	case '>':
		if world.reverse {
			world.reverse = false
			world.tick = world.space.step
		}
		world.Tick()
		world.needsDraw = time.Millisecond
		return true, nil

	case '*':
		log.Printf("solving")
		world.ui.Play()
		world.post = (*spaceWorld).checkInflection
		world.needsDraw = time.Millisecond
		return true, nil

	}
	return false, nil
}

func (world *spaceWorld) Render(g anansi.Grid, viewport image.Rectangle) {
	viewport.Min.X *= 2
	viewport.Max.X *= 2
	viewport.Min.Y *= 4
	viewport.Max.Y *= 4

	viewport.Min = viewport.Min.Add(world.b.Min)
	viewport.Max = viewport.Max.Add(world.b.Min)

	var bi anansi.Bitmap // TODO re-use
	bi.Resize(viewport.Size())
	n := 0
	for _, p := range world.p {
		if p.In(viewport) {
			bi.Set(p.Sub(viewport.Min), true)
			n++
		}
	}
	world.rendered = n
	anansi.DrawBitmap(g, bi)
	world.check()
}

func (world *spaceWorld) check() {
	if world.post != nil {
		world.post(world)
	} else {
		world.checkSeekStart()
	}
}

func (world *spaceWorld) checkSeekStart() {
	if world.rendered == 0 {
		newSize := world.b.Size()
		dsz := newSize.Sub(world.lastSize)
		if world.lastSize == image.ZP || (dsz.X < 0 && dsz.Y < 0) {
			log.Printf("viewport empty, seeking")
			world.ui.Play()
			world.post = (*spaceWorld).checkSeekStop
		}
	}
}

func (world *spaceWorld) checkSeekStop() {
	newSize := world.b.Size()
	dsz := newSize.Sub(world.lastSize)
	if world.rendered == 0 {
		if world.lastSize == image.ZP || (dsz.X <= 0 && dsz.Y <= 0) {
			return
		}
	}
	log.Printf("viewport non-empty, seek done")
	world.ui.Pause()
	world.post = nil
}

func (world *spaceWorld) checkInflection() {
	newSize := world.b.Size()
	dsz := newSize.Sub(world.lastSize)
	dt := world.t - world.lastT
	// TODO fix for reverse solve
	if dt != 0 && dsz.X/dt >= 0 && dsz.Y/dt >= 0 {
		log.Printf("solution T:%v", world.t)
		if dt > 0 {
			world.back()
		} else {
			world.space.step()
		}
		world.ui.Pause()
		world.post = nil
	}
}

type space struct {
	b    image.Rectangle
	t    int
	p, v []image.Point
}

func (sp *space) step() {
	sp.t++
	for i := range sp.p {
		sp.p[i] = sp.p[i].Add(sp.v[i])
	}
	sp.computeBounds()
}

func (sp *space) back() {
	sp.t--
	for i := range sp.p {
		sp.p[i] = sp.p[i].Sub(sp.v[i])
	}
	sp.computeBounds()
}

func (sp *space) computeBounds() {
	for i, p := range sp.p {
		if i == 0 {
			sp.b = geom.PointRect(p)
		} else {
			sp.b = sp.b.Union(geom.PointRect(p))
		}
	}
}

func (sp *space) render() anansi.Bitmap {
	var bi anansi.Bitmap
	bi.Resize(sp.b.Size())
	for _, p := range sp.p {
		bi.Set(p.Sub(sp.b.Min), true)
	}
	return bi
}

var linePat = regexp.MustCompile(
	`^position=< *(-?\d+), *(-?\d+)> +velocity=< *(-?\d+), *(-?\d+)>$`,
)

func (sp *space) load(r io.Reader) error {
	sc := bufio.NewScanner(r)
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

		var p, v image.Point
		var err error

		p.X, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid pos.X %q: %v", parts[1], err)
		}
		p.Y, err = strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf("invalid pos.X %q: %v", parts[2], err)
		}

		v.X, err = strconv.Atoi(parts[3])
		if err != nil {
			return fmt.Errorf("invalid vel.X %q: %v", parts[3], err)
		}
		v.Y, err = strconv.Atoi(parts[4])
		if err != nil {
			return fmt.Errorf("invalid vel.X %q: %v", parts[4], err)
		}

		sp.p = append(sp.p, p)
		sp.v = append(sp.v, v)
	}
	sp.computeBounds()

	return sc.Err()
}

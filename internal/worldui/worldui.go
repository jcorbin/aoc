package worldui

import (
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// UI abstracts the world ui services passed to the world being ran.
type UI interface {
	// Say places a simple message at the top of the screen, overlaid on the
	// world's rendered output.
	Say(mess string)

	// Display overlays a model message in the center of the screen, pausing
	// playback, and catching all input until <Esc> is pressed to dismiss the
	// message.
	//
	// TODO refactor this into a more general deal, with input handling.
	Display(mess string)

	Play()
	Pause()
	SetFocus(image.Point)
	ViewOffset() image.Point
	RequestRender()
}

// World is a discretely timed simulation world driven under a UI.
type World interface {
	// Init is called before the main loop commences, allowing world
	// implementations to initialize their data.
	Init(UI) error

	// Bounds returns the bounding box of the world space.
	Bounds() image.Rectangle

	// Tick should advance the simulation one step, returning false if the tick
	// wasn't complete/succesful; false return pauses playback.
	Tick() bool

	// Render the world contents to the given grid, starting from the given
	// view offset point.
	Render(g anansi.Grid, viewOffset image.Point)
}

// WorldInputHandler is an optional interface that a World implementation may
// add to add custom input processing.
type WorldInputHandler interface {
	World
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)
}

// MustRun is a convenience for implementing simulation "main".main() functions.
func MustRun(world World) {
	anansi.MustRun(Run(world))
}

// Run the given world, under a UI bound to an anansi fullscreen terminal app.
//
// Sets standard "log" package output to the Logs buffer, to prevent stderr
// logs from disrputing the fullscreen rendering.
func Run(world World) error {
	initLogs()

	in, out := os.Stdin, os.Stdout

	if !anansi.IsTerminal(in) {
		f, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	var ui worldUI
	ui.world = world

	ui.halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)
	ui.resize = anansi.Notify(syscall.SIGWINCH)
	ui.resize.Send("initialize screen size")
	ui.timer = time.NewTimer(100 * time.Second)
	ui.stopTimer()

	ui.inputHandlers = []func(e ansi.Escape, a []byte) (bool, error){
		ui.handleLowInput,
		ui.handleMessInput,
		ui.handleWorldInput,
	}

	term := anansi.NewTerm(in, out,
		&ui.halt,
		&ui.resize,
		&ui.inputReady,
		&ui.screen,
	)
	term.SetRaw(true)
	term.AddMode(
		ansi.ModeAlternateScreen,

		ansi.ModeMouseSgrExt,
		ansi.ModeMouseBtnEvent,
		// ansi.ModeMouseAnyEvent, TODO option
	)

	if err := ui.world.Init(&ui); err != nil {
		return err
	}

	if hi, ok := ui.world.(WorldInputHandler); ok {
		ui.inputHandlers = append(ui.inputHandlers, hi.HandleInput)
	}

	return term.RunWithFunc(func(term *anansi.Term) error {
		return term.Loop(&ui)
	})
}

type worldUI struct {
	halt       anansi.Signal
	resize     anansi.Signal
	inputReady anansi.InputSignal
	screen     anansi.Screen

	// TODO consider re-factoring around some component-esque abstraction over:
	// the world, logs, modal message, and the banner; maybe also component-ize
	// the playback control.
	inputHandlers []func(e ansi.Escape, a []byte) (bool, error)

	deadline time.Time
	timer    *time.Timer
	last     time.Time
	ticking  bool
	playing  bool
	playRate int // tick-per-second

	world World

	focus      image.Point
	viewOffset image.Point

	banner []byte

	mess     []byte
	messSize image.Point
}

func (ui *worldUI) Play() {
	ui.ticking = false
	ui.playing = false
}

func (ui *worldUI) Pause() {
	ui.ticking = false
	ui.playing = false
}

func (ui *worldUI) SetFocus(p image.Point) {
	ui.focus = p
}

func (ui *worldUI) ViewOffset() image.Point {
	return ui.viewOffset
}

func (ui *worldUI) RequestRender() {
	ui.setTimer(5 * time.Millisecond)
}

func (ui *worldUI) setTimer(d time.Duration) {
	deadline := time.Now().Add(d)
	if !ui.deadline.IsZero() && deadline.After(ui.deadline) {
		return
	}

	if ui.timer == nil {
		ui.timer = time.NewTimer(d)
	} else {
		ui.timer.Reset(d)
	}
	ui.deadline = deadline
}

func (ui *worldUI) stopTimer() {
	ui.timer.Stop()
	ui.deadline = time.Time{}
	select {
	case <-ui.timer.C:
	default:
	}
}

func (ui *worldUI) Update(term *anansi.Term) (redraw bool, _ error) {
	select {
	case sig := <-ui.halt.C:
		return false, anansi.SigErr(sig)

	case <-ui.resize.C:
		if err := ui.screen.SizeToTerm(term); err != nil {
			return false, err
		}
		ui.RequestRender()

	case <-ui.inputReady.C:
		_, err := term.ReadAny()
		herr := ui.handleInput(term)
		if err == nil {
			err = herr
		}
		if err != nil {
			return false, err
		}

	case now := <-ui.timer.C:
		ui.deadline = time.Time{}
		ui.advance(now)
		redraw = true
	}
	return redraw, nil
}

func (ui *worldUI) advance(now time.Time) {
	// no updates while displaying a message
	if !ui.ticking {
		ui.last = now
		return
	}

	// single-step
	if !ui.playing {
		ui.world.Tick()
		ui.ticking = false
		ui.last = now
		return
	}

	// advance playback
	if ticks := int(math.Round(float64(now.Sub(ui.last)) / float64(time.Second) * float64(ui.playRate))); ticks > 0 {
		const maxTicks = 100000
		if ticks > maxTicks {
			ticks = maxTicks
		}
		for i := 0; i < ticks; i++ {
			if !ui.world.Tick() {
				ui.playing = false
				break
			}
		}
		ui.last = now
	}

	ui.ticking = true
	ui.setTimer(10 * time.Millisecond) // TODO compute next time when ticks > 0; avoid spurious wakeup
}

func (ui *worldUI) Say(mess string) {
	ui.banner = []byte(mess)
}

func (ui *worldUI) Display(mess string) {
	if mess == "" {
		ui.mess = nil
		ui.messSize = image.ZP
	} else {
		ui.mess = []byte(mess)
		ui.messSize = measureTextBox(ui.mess).Size()
	}
	ui.RequestRender()
}

func (ui *worldUI) handleInput(term *anansi.Term) error {
	defer func(before int) {
		after := Logs.Len()
		if after-before > 0 {
			ui.RequestRender()
		}
	}(Logs.Len())
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		var handled bool
		var err error
		for i := 0; i < len(ui.inputHandlers); i++ {
			if handled, err = ui.inputHandlers[i](e, a); err != nil {
				return err
			}
			if handled {
				break
			}
		}
	}
	return nil
}

func (ui *worldUI) handleLowInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	case 0x03: // stop on Ctrl-C
		return true, fmt.Errorf("read %v", e)

	case 0x0c: // clear screen on Ctrl-L
		ui.screen.Clear()           // clear virtual contents
		ui.screen.To(ansi.Pt(1, 1)) // cursor back to top
		ui.screen.Invalidate()      // force full redraw
		ui.RequestRender()
		return true, nil

	}
	return false, nil
}

func (ui *worldUI) handleMessInput(e ansi.Escape, a []byte) (bool, error) {
	// no message, ignore
	if ui.mess == nil {
		return false, nil
	}

	switch e {

	// <Esc> to dismiss message
	case ansi.Escape('\x1b'):
		ui.Display("")
		return true, nil

	// eat any other input when a message is shown
	default:
		return true, nil
	}
}

func (ui *worldUI) handleWorldInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {
	// arrow keys to move view
	case ansi.CUB, ansi.CUF, ansi.CUU, ansi.CUD:
		if d, ok := ansi.DecodeCursorCardinal(e, a); ok {
			p := ui.focus.Add(d)
			bounds := ui.world.Bounds()
			if p.X < bounds.Min.X {
				p.X = bounds.Min.X
			}
			if p.Y < bounds.Min.Y {
				p.Y = bounds.Min.Y
			}
			if p.X >= bounds.Max.X {
				p.X = bounds.Max.X - 1
			}
			if p.Y >= bounds.Max.Y {
				p.Y = bounds.Max.Y - 1
			}
			if ui.focus != p {
				ui.focus = p
				ui.RequestRender()
			}
		}
		return true, nil

	// step
	case ansi.Escape('.'):
		ui.RequestRender()
		ui.ticking = true
		return true, nil

	// play/pause
	case ansi.Escape(' '):
		ui.playing = !ui.playing
		if !ui.playing {
			ui.stopTimer()
			log.Printf("pause")
		} else {
			ui.last = time.Now()
			if ui.playRate == 0 {
				ui.playRate = 1
			}
			ui.ticking = true
			log.Printf("play at %v ticks/s", ui.playRate)
		}
		ui.RequestRender()
		return true, nil

	// speed control
	case ansi.Escape('+'):
		ui.playRate *= 2
		log.Printf("speed up to %v ticks/s", ui.playRate)
		ui.RequestRender()
		return true, nil
	case ansi.Escape('-'):
		rate := ui.playRate / 2
		if rate <= 0 {
			rate = 1
		}
		if ui.playRate != rate {
			ui.playRate = rate
			log.Printf("slow down to %v ticks/s", ui.playRate)
		}
		ui.RequestRender()
		return true, nil

	}

	return false, nil
}

func (ui *worldUI) WriteTo(w io.Writer) (n int64, err error) {
	ui.viewOffset = ui.screen.Bounds().Size().Div(2).Sub(ui.focus)

	ui.screen.Clear()
	ui.world.Render(ui.screen.Grid, ui.viewOffset)
	DrawLogs(ui.screen.Grid.SubAt(ansi.Pt(
		1, ui.screen.Bounds().Dy()-5,
	)), true) // TODO support top-align option
	ui.overlayBanner()
	ui.overlayMess()
	return ui.screen.WriteTo(w)
}

func (ui *worldUI) overlayBanner() {
	at := ui.screen.Grid.Rect.Min
	bannerWidth := measureTextBox(ui.banner).Dx()
	screenWidth := ui.screen.Bounds().Dx()
	at.X += screenWidth/2 - bannerWidth/2
	writeIntoGrid(ui.screen.Grid.SubAt(at), ui.banner)
}

func (ui *worldUI) overlayMess() {
	if ui.mess == nil || ui.messSize == image.ZP {
		return
	}
	screenSize := ui.screen.Bounds().Size()
	screenMid := screenSize.Div(2)
	messMid := ui.messSize.Div(2)
	offset := screenMid.Sub(messMid)
	writeIntoGrid(ui.screen.Grid.SubAt(ui.screen.Grid.Rect.Min.Add(offset)), ui.mess)
}

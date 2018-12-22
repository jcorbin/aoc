package layerui

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// LayerApp implements an anansi.Loop around a list of layers.
type LayerApp struct {
	Layer
	now        time.Time
	halt       anansi.Signal
	resize     anansi.Signal
	inputReady anansi.InputSignal
	screen     anansi.Screen
	timer      anansi.Timer
}

// Run runs the given layer under a new anansi terminal attached to
// os.Stdin/out and setup to process signals. The args may be additional
// ansi.Modes to set or anansi.Context pieces to attach to the terminal.
func Run(args ...interface{}) error {
	var app LayerApp

	in, out, err := OpenTermFiles(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	ctx := anansi.Contexts(
		&app.halt,
		&app.resize,
		&app.inputReady,
		&app.screen,
	)

	modes := []ansi.Mode{
		ansi.ModeAlternateScreen,
	}

	for i := range args {
		switch arg := args[i].(type) {
		case anansi.Context:
			ctx = anansi.Contexts(ctx, arg)
		case ansi.Mode:
			modes = append(modes, arg)
		case Layer:
			app.Layer = Layers(app.Layer, arg)
		default:
			panic(fmt.Sprintf("unsupported LayerApp.Run argument type %T", arg))
		}
	}

	app.SetupSignals()

	term := anansi.NewTerm(in, out, ctx)
	term.SetRaw(true)
	term.SetEcho(false)
	term.AddMode(modes...)

	// TODO consider borging anansi.Loop* into here now that it's the
	// primary/only user.
	return term.RunWithFunc(func(term *anansi.Term) error {
		return term.Loop(&app)
	})
}

// SetupSignals enables halt and resize signal notification.
func (app *LayerApp) SetupSignals() {
	app.halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)
	app.resize = anansi.Notify(syscall.SIGWINCH)
	app.resize.Send("initialize screen size")
}

// Update waits for a signal, input, or draw timer to fire, dispatching to the
// layers accordingly:
//
// On input, Layer.HandleInput is called in order, stopping on the first one
// that handles it or returns an error. While dispatching input,
// Layer.NeedsDraw is polled, retaining any minimal non-zero duration; if
// non-zero, a draw timer is set.
//
// When the draw timer fires, or when the terminal is resized, the terminal is
// flushed using the LayerApp.WriteTo.
func (app *LayerApp) Update(term *anansi.Term) (bool, error) {
	select {
	case sig := <-app.halt.C:
		return false, anansi.SigErr(sig)

	case <-app.resize.C:
		err := app.screen.SizeToTerm(term)
		if err == nil {
			app.timer.Request(5 * time.Millisecond)
		}
		return false, err

	case <-app.inputReady.C:
		_, err := term.ReadAny()
		herr := app.handleInput(term)
		if err == nil {
			err = herr
		}
		return false, err

	case now := <-app.timer.C:
		app.now = now
		return true, nil

	default:
		return false, nil
	}
}

func (app *LayerApp) setTimerIfNeeded() {
	if d := app.NeedsDraw(); d > 0 {
		app.timer.Request(d)
	}
}

func (app *LayerApp) handleInput(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		if handled, err := app.handleLowInput(e, a); err != nil {
			return err
		} else if !handled {
			handled, err = app.HandleInput(e, a)
			if err != nil {
				return err
			}
		}
	}
	app.setTimerIfNeeded()
	return nil
}

func (app *LayerApp) handleLowInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	case 0x03: // stop on Ctrl-C
		return true, fmt.Errorf("read %v", e)

	case 0x0c: // clear screen on Ctrl-L
		app.screen.Clear()           // clear virtual contents
		app.screen.To(ansi.Pt(1, 1)) // cursor back to top
		app.screen.Invalidate()      // force full redraw
		app.timer.Request(5 * time.Millisecond)
		return true, nil

	}
	return false, nil
}

// WriteTo draws the layer to a newly cleared virtual screen, then
// flushes it to the terminal.
func (app *LayerApp) WriteTo(w io.Writer) (n int64, err error) {
	app.screen.Clear()
	app.Draw(app.screen, app.now)
	if n, err = app.screen.WriteTo(w); err == nil {
		app.setTimerIfNeeded()
	}
	return n, err
}

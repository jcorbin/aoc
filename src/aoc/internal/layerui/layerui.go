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

// Layer is a composable user interface element.
type Layer interface {
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)
	Draw(screen anansi.Screen, now time.Time)
	NeedsDraw() time.Duration
}

// Layers combines the given layer(s) into a single layer that dispatches
// HandleInput in order and Draw in reverse order. Its NeedsDraw method returns
// the smallest non-zero value from the constituent layers.
func Layers(ls ...Layer) Layer {
	if len(ls) == 0 {
		return nil
	}
	a := ls[0]
	for i := 1; i < len(ls); i++ {
		b := ls[i]
		if b == nil || b == Layer(nil) {
			continue
		}
		if a == nil || a == Layer(nil) {
			a = b
			continue
		}
		as, haveAs := a.(layers)
		bs, haveBs := b.(layers)
		if haveAs && haveBs {
			a = append(as, bs...)
		} else if haveAs {
			a = append(as, b)
		} else if haveBs {
			a = append(layers{a}, bs...)
		} else {
			a = layers{a, b}
		}
	}
	return a
}

type layers []Layer

func (ls layers) NeedsDraw() (d time.Duration) {
	for i := 0; i < len(ls); i++ {
		nd := ls[i].NeedsDraw()
		if d == 0 || (nd > 0 && nd < d) {
			d = nd
		}
	}
	return d
}

func (ls layers) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	for i := 0; i < len(ls); i++ {
		if handled, err = ls[i].HandleInput(e, a); handled || err != nil {
			return handled, err
		}
	}
	return false, nil
}

func (ls layers) Draw(screen anansi.Screen, now time.Time) {
	for i := len(ls) - 1; i >= 0; i-- {
		ls[i].Draw(screen, now)
	}
}

// LayerUI implements an anansi.Loop around a list of layers.
type LayerUI struct {
	Layer
	now        time.Time
	halt       anansi.Signal
	resize     anansi.Signal
	inputReady anansi.InputSignal
	screen     anansi.Screen
	timer      anansi.Timer
}

// Run a new LayerUI.
func Run(args ...interface{}) error {
	var lui LayerUI
	in, out, err := OpenTermFiles(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	ctx := anansi.Contexts(
		&lui.halt,
		&lui.resize,
		&lui.inputReady,
		&lui.screen,
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
			lui.Layer = Layers(lui.Layer, arg)
		default:
			panic(fmt.Sprintf("unsupported LayerUI.Run argument type %T", arg))
		}
	}

	lui.SetupSignals()

	term := anansi.NewTerm(in, out, ctx)
	term.SetRaw(true)
	term.SetEcho(false)
	term.AddMode(modes...)

	// TODO consider borging anansi.Loop* into here now that it's the
	// primary/only user.
	return term.RunWithFunc(func(term *anansi.Term) error {
		return term.Loop(&lui)
	})
}

// SetupSignals enables halt and resize signal notification.
func (lui *LayerUI) SetupSignals() {
	lui.halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)
	lui.resize = anansi.Notify(syscall.SIGWINCH)
	lui.resize.Send("initialize screen size")
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
// flushed using the LayerUI.WriteTo.
func (lui *LayerUI) Update(term *anansi.Term) (bool, error) {
	select {
	case sig := <-lui.halt.C:
		return false, anansi.SigErr(sig)

	case <-lui.resize.C:
		err := lui.screen.SizeToTerm(term)
		if err == nil {
			lui.timer.Request(5 * time.Millisecond)
		}
		return false, err

	case <-lui.inputReady.C:
		_, err := term.ReadAny()
		herr := lui.handleInput(term)
		if err == nil {
			err = herr
		}
		return false, err

	case now := <-lui.timer.C:
		lui.now = now
		return true, nil

	default:
		return false, nil
	}
}

func (lui *LayerUI) setTimerIfNeeded() (d time.Duration) {
	d = lui.NeedsDraw()
	if d > 0 {
		lui.timer.Request(d)
	}
	return d
}

func (lui *LayerUI) handleInput(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		handled, err := lui.handleLowInput(e, a)
		if err == nil && !handled {
			handled, err = lui.HandleInput(e, a)
		}
		if err != nil {
			return err
		}
	}
	lui.setTimerIfNeeded()
	return nil
}

func (lui *LayerUI) handleLowInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	case 0x03: // stop on Ctrl-C
		return true, fmt.Errorf("read %v", e)

	case 0x0c: // clear screen on Ctrl-L
		lui.screen.Clear()           // clear virtual contents
		lui.screen.To(ansi.Pt(1, 1)) // cursor back to top
		lui.screen.Invalidate()      // force full redraw
		lui.timer.Request(5 * time.Millisecond)
		return true, nil

	}
	return false, nil
}

// WriteTo calls Layer.Draw in in reverse order.
func (lui *LayerUI) WriteTo(w io.Writer) (n int64, err error) {
	lui.screen.Clear()
	lui.Draw(lui.screen, lui.now)
	if n, err = lui.screen.WriteTo(w); err == nil {
		lui.setTimerIfNeeded()
	}
	return n, err
}

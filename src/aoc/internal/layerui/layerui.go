package layerui

import (
	"fmt"
	"io"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// Layer is a composable user interface element run under LayerUI.
type Layer interface {
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)
	Draw(anansi.Screen, time.Time)
	NeedsDraw() time.Duration
}

// LayerUI implements an anansi.Loop around a list of Layers.
type LayerUI struct {
	Layers []Layer

	needsDraw  time.Duration
	halt       anansi.Signal
	resize     anansi.Signal
	inputReady anansi.InputSignal
	screen     anansi.Screen
	timer      drawTimer
}

// Enter registers signal notification, input notification, and initializes
// the screen.
func (lui *LayerUI) Enter(term *anansi.Term) error {
	if err := lui.halt.Enter(term); err != nil {
		return err
	}
	if err := lui.resize.Enter(term); err != nil {
		return err
	}
	if err := lui.inputReady.Enter(term); err != nil {
		return err
	}
	if err := lui.screen.Enter(term); err != nil {
		return err
	}
	return nil
}

// Exit is the converse of Enter.
func (lui *LayerUI) Exit(term *anansi.Term) error {
	err := lui.screen.Enter(term)
	if eerr := lui.inputReady.Enter(term); err == nil {
		err = eerr
	}
	if eerr := lui.resize.Enter(term); err == nil {
		err = eerr
	}
	if eerr := lui.halt.Enter(term); err == nil {
		err = eerr
	}
	return err
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
			lui.timer.set(5 * time.Millisecond)
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
		lui.timer.update(now)
		return true, nil

	default:
		return false, nil
	}
}

func (lui *LayerUI) setTimerIfNeeded() time.Duration {
	needsDraw := lui.needsDraw
	for i := 0; i < len(lui.Layers); i++ {
		nd := lui.Layers[i].NeedsDraw()
		if needsDraw == 0 || (nd > 0 && nd < needsDraw) {
			needsDraw = nd
		}
	}
	if needsDraw > 0 {
		lui.timer.set(needsDraw)
	}
	return needsDraw
}

func (lui *LayerUI) handleInput(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		handled, err := lui.handleLowInput(e, a)
		if err != nil {
			return err
		}
		if handled {
			continue
		}

		for i := 0; i < len(lui.Layers); i++ {
			if handled, err = lui.Layers[i].HandleInput(e, a); err != nil {
				return err
			}
			if handled {
				break
			}
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
		lui.needsDraw = 5 * time.Millisecond
		return true, nil

	}
	return false, nil
}

// WriteTo calls Layer.Draw in in reverse order.
func (lui *LayerUI) WriteTo(w io.Writer) (n int64, err error) {
	lui.needsDraw = 0
	lui.screen.Clear()
	for i := len(lui.Layers) - 1; i >= 0; i-- {
		lui.Layers[i].Draw(lui.screen, lui.timer.last)
	}
	if n, err = lui.screen.WriteTo(w); err == nil {
		lui.setTimerIfNeeded()
	}
	return n, err
}

type drawTimer struct {
	C        <-chan time.Time
	deadline time.Time
	last     time.Time
	timer    *time.Timer
}

func (dt *drawTimer) update(now time.Time) {
	dt.deadline = time.Time{}
	dt.last = now
}

func (dt *drawTimer) set(d time.Duration) {
	deadline := time.Now().Add(d)
	if !dt.deadline.IsZero() && deadline.After(dt.deadline) {
		return
	}
	if dt.timer == nil {
		dt.timer = time.NewTimer(d)
		dt.C = dt.timer.C
	} else {
		dt.timer.Reset(d)
	}
	dt.deadline = deadline
}

func (dt *drawTimer) stop() {
	dt.timer.Stop()
	dt.deadline = time.Time{}
	select {
	case <-dt.C:
	default:
	}
}
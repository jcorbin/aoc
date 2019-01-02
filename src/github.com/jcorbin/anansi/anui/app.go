package anui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// LayerApp supports running a Layer under an anansi.Term, by implementing
// anansi.Runner. Most users should call the toplevel Run function, rather than
// building a LayerApp directly.
type LayerApp struct {
	Layer
	Halt       anansi.Signal
	Resize     anansi.Signal
	InputReady anansi.InputSignal
	Timer      anansi.Timer
	Screen     anansi.Screen

	audit auditer
}

// Run a fresh LayerApp under os.Stdin/os.Stdout, with SIGTERM, SIGINT, and
// SIGWINCH handling appropriate for a nolmal fullscreen terminal application.
//
// See LayerApp.Build for details.
func Run(args ...interface{}) error {
	var app LayerApp
	in, out, err := OpenTermFiles(os.Stdin, os.Stdout) // TODO option-ize... or just promote to Run(in, out, args...)
	if err != nil {
		return err
	}
	app.Halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT) // TODO option-ize
	app.Resize = anansi.Notify(syscall.SIGWINCH)              // TODO option-ize
	term, err := app.Build(in, out, args...)
	if err == nil {
		err = term.RunWith(&app)
	}
	return err
}

// Build a new anansi.Term attached to the given files, and the app's loop
// signaling.
//
// Each of the args may be:
//     - a Layer to add to the app, using Layers
//     - an ansi.Mode to add to the terminal
//     - or an anansi.Context to run under the term.
func (app *LayerApp) Build(in, out *os.File, args ...interface{}) (*anansi.Term, error) {
	ctx := anansi.Contexts(
		&app.Halt,
		&app.Resize,
		&app.InputReady,
		&app.Screen,
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

	if err := app.audit.Open("screen_audit.log"); err != nil {
		return nil, err
	}

	term := anansi.NewTerm(in, out, ctx)
	term.SetRaw(true)
	term.SetEcho(false)
	term.AddMode(modes...)

	return term, nil
}

// Run the app's event harndling loop.
//
// The loop stops on halt signal, and resizes the app's virtual screen to the
// terminal's size when signaled. After resize the draw timer is set to fire
// as soon as possible.
//
// When input is ready, reads any available input from the terminal
// (non-blocking mode), and then processes it. After processing some low level
// input, like Ctrl-C, input is passed to the app Layer.
//
// When the draw timer fires, the app's virtual screen is cleared, the Layer
// Draw()n to it, and then the app is flushed to the terminal.
func (app *LayerApp) Run(term *anansi.Term) error {
	app.Resize.Send("initialize screen size")
	for {
		select {

		case sig := <-app.Halt.C:
			return anansi.SigErr(sig)

		case <-app.Resize.C:
			if err := app.Screen.SizeToTerm(term); err != nil {
				return err
			}
			app.audit.screen.Resize(app.Screen.Bounds().Size())
			app.Timer.Request(time.Millisecond)

		case <-app.InputReady.C:
			_, err := term.ReadAny()
			if herr := app.handleInput(term); err == nil {
				err = herr
			}
			if err != nil {
				return err
			}

		case now := <-app.Timer.C:
			app.Screen.Clear()
			app.Draw(app.Screen, now)
			if err := term.Flush(app); err != nil {
				return err
			}

		}
	}
}

func (app *LayerApp) setTimerIfNeeded() {
	const minTimeout = time.Second / 120
	if d := app.NeedsDraw(); d > 0 {
		if d < minTimeout {
			d = minTimeout
		}
		app.Timer.Request(d)
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
		app.Screen.Clear()           // clear virtual contents
		app.Screen.To(ansi.Pt(1, 1)) // cursor back to top
		app.Screen.Invalidate()      // force full redraw
		app.Timer.Request(time.Millisecond)
		app.audit.reset(&app.Screen)
		return true, nil

	}
	return false, nil
}

// WriteTo writes the app's virtual screen to the given io.Writer, setting any
// next needed timer if the write succeeds.
func (app *LayerApp) WriteTo(w io.Writer) (n int64, err error) {
	if app.audit.logFile != nil {
		var fin func(error)
		w, fin = app.audit.audit(&app.Screen, w)
		defer fin(err)
	}

	n, err = app.Screen.WriteTo(w)
	if err == nil {
		app.setTimerIfNeeded()
	}
	return n, err
}

type auditer struct {
	buf    bytes.Buffer
	screen anansi.ScreenState

	lastAudit bool

	logBuf  bytes.Buffer
	logFile *os.File
	err     error
}

func (aud *auditer) reset(sc *anansi.Screen) {
	aud.screen.Clear()
	aud.screen.To(ansi.Pt(1, 1))
	aud.screen.Resize(sc.Bounds().Size())
	aud.lastAudit = true
}

func (aud *auditer) audit(sc *anansi.Screen, w io.Writer) (io.Writer, func(error)) {
	w = io.MultiWriter(w, &aud.buf)
	prior := sc.Real
	prior.Rune = append([]rune(nil), prior.Rune...)
	prior.Attr = append([]ansi.SGRAttr(nil), prior.Attr...)
	return w, func(err error) {
		aud.fin(prior, sc.ScreenState, err)
	}
}

// AuditRecord FIXME
type AuditRecord struct {
	Ok      bool
	Prior   anansi.ScreenState
	Virtual anansi.ScreenState
	Update  string
	Audit   anansi.ScreenState
}

func (aud *auditer) fin(prior, virt anansi.ScreenState, err error) {
	if err != nil {
		if unwrapOSError(err) == syscall.EWOULDBLOCK {
			aud.Printf("audit smeared over %v", err)
		} else {
			aud.Printf("audit write failed: %v", err)
			aud.buf.Reset()
		}
		return
	}

	update := aud.buf.Bytes()
	anansi.ProcessBytes(&aud.screen, update) // XXX is broken?
	auditOK := virt.Grid.Eq(aud.screen.Grid, ' ')
	// if auditOK != aud.lastAudit { }

	if dat, err := json.Marshal(AuditRecord{
		Ok:      auditOK,
		Prior:   prior,
		Virtual: virt,
		Update:  string(update),
		Audit:   aud.screen,
	}); err != nil {
		aud.Printf("ERROR marshaling audit data: %v", err)
	} else {
		aud.Printf("screen diff audit ok:%v %s", auditOK, dat)
	}

	aud.lastAudit = auditOK
	aud.buf.Reset()
}

func (aud *auditer) Open(name string) error {
	aud.Close()
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	aud.logFile = f
	aud.logBuf.Grow(64 * 1024)
	aud.err = nil
	log.Printf("auditing screen diffs to %q", aud.logFile.Name())
	return nil
}

func (aud *auditer) Close() error {
	if aud.logFile == nil {
		return nil
	}
	err := aud.logFile.Close()
	aud.logFile = nil
	return err
}

func (aud *auditer) Printf(mess string, args ...interface{}) {
	if aud.logFile == nil || aud.err != nil {
		return
	}
	aud.logBuf.WriteString(time.Now().Format(time.RFC3339Nano))
	aud.logBuf.WriteRune(' ')
	fmt.Fprintf(&aud.logBuf, mess, args...)
	aud.logBuf.WriteRune('\n')
	_, aud.err = aud.logBuf.WriteTo(aud.logFile)
}

func unwrapOSError(err error) error {
	for {
		switch val := err.(type) {
		case *os.PathError:
			err = val.Err
		case *os.LinkError:
			err = val.Err
		case *os.SyscallError:
			err = val.Err
		default:
			return err
		}
	}
}

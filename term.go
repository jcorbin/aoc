package anansi

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// NewTerm constructs a new terminal attached the given file pair, and with the
// given context.
func NewTerm(in, out *os.File, cs ...Context) *Term {
	term := &Term{}
	term.Input.File = in
	term.Output.File = out
	term.ctx = Contexts(
		&term.Input,
		&term.Output,
		&term.Attr,
		&term.Mode,
		Contexts(cs...))
	return term
}

// Term combines a terminal file handle with attribute control and further
// Context-ual state.
type Term struct {
	Attr
	Mode
	Input
	Output

	active bool
	under  bool
	ctx    Context
}

// RunWith runs the given runner within the terminal's context, Enter()ing it
// if necessary, and Exit()ing it if Enter() was called after the given runner
// returns. Exit() is called even if the within runner returns an error or
// panics.
//
// If the context implements a `Close() error` method, then it will also be
// called immediately after Exit(). This allows a Context implementation to
// differentiate between temporary teardown, e.g. suspending under RunWithout,
// and final teardown as RunWith returns.
func (term *Term) RunWith(runner Runner) (err error) {
	if term.active {
		return runner.Run(term)
	}

	term.active = true
	defer func() {
		term.active = false
	}()

	if !term.under {
		term.under = true
		defer func() {
			term.under = false
		}()
	}

	if term.ctx == nil {
		term.ctx = Contexts(&term.Attr, &term.Mode)
	}

	if cl, ok := term.ctx.(interface{ Close() error }); ok {
		defer func() {
			if cerr := cl.Close(); err == nil {
				err = cerr
			}
		}()
	}

	defer func() {
		if cerr := term.ctx.Exit(term); err == nil {
			err = cerr
		}
	}()
	if err = term.ctx.Enter(term); err == nil {
		err = runner.Run(term)
	}
	return err
}

// RunWithout runs the given runner without the terminal's context, Exit()ing
// it if necessary, and Enter()ing it if deactivation was necessary.
// Re-Enter() is not called is not done if a non-nil error is returned, or if
// the without runner panics.
func (term *Term) RunWithout(runner Runner) (err error) {
	if !term.active {
		return runner.Run(term)
	}
	if err = term.ctx.Exit(term); err == nil {
		term.active = false
		if err = runner.Run(term); err == nil {
			if err = term.ctx.Enter(term); err == nil {
				term.active = true
			}
		}
	}
	return err
}

// RunWithFunc is a convenience for RunWith around a function.
func (term *Term) RunWithFunc(f func(*Term) error) error {
	return term.RunWith(RunnerFunc(f))
}

// RunWithoutFunc is a convenience for RunWithout around a function.
func (term *Term) RunWithoutFunc(f func(*Term) error) error {
	return term.RunWithout(RunnerFunc(f))
}

// Runner runs under term.RunWith or term.RunWithout.
type Runner interface {
	Run(*Term) error
}

// RunnerFunc is a convenience for implementing Runner.
type RunnerFunc func(*Term) error

// Run calls the function.
func (f RunnerFunc) Run(term *Term) error { return f(term) }

// Suspend signals the process to stop, and blocks on its later restart. If the
// terminal is currently active, this is done under RunWithout to restore prior
// terminal state.
func (term *Term) Suspend() error {
	if term.active {
		return term.RunWithoutFunc((*Term).Suspend)
	}

	cont := make(chan os.Signal)
	signal.Notify(cont, syscall.SIGCONT)
	defer signal.Stop(cont)
	log.Printf("suspending")
	if err := syscall.Kill(0, syscall.SIGTSTP); err != nil {
		return err
	}
	sig := <-cont
	log.Printf("resumed, signal: %v", sig)
	return nil
}

// ExitError may be implemented by an error to customize the exit code under
// MustRun.
type ExitError interface {
	error
	ExitCode() int
}

// MustRun is a useful wrapper for the outermost Term.RunWith: if the error
// value implements ExitError, and its ExitCode method returns non-0, it calls
// os.Exit; otherwise any non-nil error value is log.Fatal-ed.
func MustRun(err error) {
	if err != nil {
		if ex, ok := err.(ExitError); ok {
			log.Printf("exiting due to %v", ex)
			if ec := ex.ExitCode(); ec != 0 {
				os.Exit(ec)
			}
		}
		log.Fatalln(err)
	}
}

package anui

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// Logs is an in-memory buffer of all logs written through the standard "log"
// package.
var Logs bytes.Buffer

var logsSetup bool

// WithOpenLogFile is a convenience that calls the given function after
// OpenLogFile, restoring os.Stderr log output before returning any error.
func WithOpenLogFile(name string, f func() error) error {
	if name == "" {
		InitLogs()
	} else if err := OpenLogFile(name); err != nil {
		return err
	}
	defer log.SetOutput(os.Stderr)
	return f()
}

// OpenLogFile creates a file with the given name, and sets the "log" package
// output to be an io.MultiWriter to it and the Logs buffer.
func OpenLogFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create logfile %q: %v", name, err)
	}
	log.SetOutput(io.MultiWriter(
		&Logs,
		f,
	))
	logsSetup = true
	return nil
}

// InitLogs installes the Logs buffer by calling log.SetOutput, unless
// OpenLogFile or a prior InitLogs has already done so.
func InitLogs() {
	if !logsSetup {
		log.SetOutput(&Logs)
		logsSetup = true
	}
}

// TODO cap the buffer, load from file if scroll past..

// LogLayer implements a layer for displaying in-memory buffered Logs.
type LogLayer struct {
	SubGrid func(g anansi.Grid, numLines int) anansi.Grid
	lastLen int
}

//HandleInput is a no-op.
func (ll LogLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	// TODO support scrolling
	return false, nil
}

// Draw overlays the tail of buffered Logs content into the screen grid. If
// LogLayer.SubGrid is not nil, it is used to select a sub-grid of the screen.
func (ll *LogLayer) Draw(screen anansi.Screen, now time.Time) {
	lb := Logs.Bytes()
	numLines := bytes.Count(lb, []byte("\n"))
	// if len(lb) > 0 { numLines++ }

	g := screen.Grid
	if ll.SubGrid != nil {
		g = ll.SubGrid(g, numLines)
	}

	height := g.Bounds().Dy()

	off := len(lb)
	for i := 0; i < height; i++ {
		b := lb[:off]
		i := bytes.LastIndexByte(b, '\n')
		if i < 0 {
			off = 0
			break
		}
		off -= len(b) - i
	}
	for off < len(lb) && lb[off] == '\n' {
		off++
	}

	WriteIntoGrid(g, lb[off:])

	ll.lastLen = len(lb)
}

// NeedsDraw returns non-zero if more logs have been written since last Draw.
func (ll LogLayer) NeedsDraw() time.Duration {
	if Logs.Len() > ll.lastLen {
		return time.Millisecond
	}
	return 0
}

// BottomNLines returns a function that returns a bottom-aligned subgrid of at
// most n lines within the grid it's passed.
func BottomNLines(n int) func(g anansi.Grid, numLines int) anansi.Grid {
	return func(g anansi.Grid, numLines int) anansi.Grid {
		if numLines > n {
			numLines = n
		}
		return g.SubAt(ansi.Pt(
			1, g.Bounds().Dy()-numLines,
		))
	}
}

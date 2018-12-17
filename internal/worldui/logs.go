package worldui

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jcorbin/anansi"
)

// Logs is an in-memory buffer of all logs written through the standard "log"
// package.
var Logs bytes.Buffer

var logsSetup bool

// MustOpenLogFile is a convenience that prints to stderr and exits if
// OpenLogFile fails.
func MustOpenLogFile(name string) {
	if name == "" {
		initLogs()
		return
	}
	if err := OpenLogFile(name); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
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

func initLogs() {
	if !logsSetup {
		log.SetOutput(&Logs)
		logsSetup = true
	}
}

// TODO support scrolling, cap the buffer, load from file if scroll past..

// DrawLogs writes buffered Logs content into the given screen grid.  If
// bottomAlign is true than the grid may be shrunk by moving it's top point
// down to fit the number of lines.
func DrawLogs(g anansi.Grid, bottomAlign bool) {
	height := g.Bounds().Dy()
	lb := Logs.Bytes()

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

	b := lb[off:]

	numLines := bytes.Count(b, []byte("\n"))
	if len(b) > 0 {
		numLines++
	}

	if bottomAlign && numLines < height {
		pt := g.Rect.Min
		pt.Y += height - numLines
		g = g.SubAt(pt)
	}

	writeIntoGrid(g, b)
}

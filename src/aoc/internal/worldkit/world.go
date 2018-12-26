package worldkit

import (
	"time"

	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/anui"
)

// World is a discretely timed simulation world driven under a UI.
type World interface {
	// Tick should advance the simulation one step, returning false if the tick
	// wasn't complete/succesful; false return pauses playback.
	Tick() bool

	NeedsDraw() time.Duration
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)

	anui.ViewClient
}

package layerui

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// World is a discretely timed simulation world driven under a UI.
type World interface {
	// Tick should advance the simulation one step, returning false if the tick
	// wasn't complete/succesful; false return pauses playback.
	Tick() bool

	NeedsDraw() time.Duration
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)

	ViewClient
}

// WorldLayer implements a layer that controls and displays a World
// simulation.
type WorldLayer struct {
	World
	View Layer

	last     time.Time
	ticking  bool
	playing  bool
	playRate int // tick-per-second

	needsDraw time.Duration
}

// Play starts playback.
func (world *WorldLayer) Play() {
	world.ticking = true
	world.playing = true
	world.needsDraw = time.Millisecond
}

// Pause stops playback.
func (world *WorldLayer) Pause() {
	world.ticking = false
	world.playing = false
	world.needsDraw = time.Millisecond
}

// Update advances the world simulation by calling tick one or more times if
// enabled.
func (world *WorldLayer) Update(now time.Time) {
	world.needsDraw = 0

	// no updates while displaying a message
	if !world.ticking {
		world.last = now
		return
	}

	// single-step
	if !world.playing {
		world.Tick()
		world.ticking = false
		world.last = now
		return
	}

	// advance playback
	if ticks := int(math.Round(float64(now.Sub(world.last)) / float64(time.Second) * float64(world.playRate))); ticks > 0 {
		const maxTicks = 100000
		if ticks > maxTicks {
			ticks = maxTicks
		}
		for i := 0; world.playing && world.ticking && i < ticks; i++ {
			if !world.Tick() {
				world.playing = false
				break
			}
		}
		world.last = now
	}

	world.ticking = true
	if world.playing {
		world.needsDraw = world.untilNextTick()
		if world.needsDraw < time.Second/60 {
			world.needsDraw = time.Second / 60
		}
	}
}

func (world *WorldLayer) untilNextTick() time.Duration {
	if world.playRate == 0 {
		world.playRate = 1
	}
	return time.Second / time.Duration(world.playRate)
}

func (world *WorldLayer) init() {
	if world.View == nil {
		world.View = DrawFuncLayer(world.DrawWorld)
	}
}

// NeedsDraw returns non-zero if the layer needs to be drawn.
func (world *WorldLayer) NeedsDraw() time.Duration {
	world.init()
	return minNeedsDraw(
		world.needsDraw,
		world.World.NeedsDraw(),
		world.View.NeedsDraw(),
	)
}

// HandleInput handles world input: cursor keys to move view focus, '.' to
// step, ' ' to play/pause, and '+'/'-' for playback rate control.
func (world *WorldLayer) HandleInput(e ansi.Escape, a []byte) (bool, error) {
	world.init()

	switch e {
	// step
	case ansi.Escape('.'):
		world.needsDraw = time.Millisecond
		world.ticking = true
		return true, nil

	// play/pause
	case ansi.Escape(' '):
		world.last = time.Now()
		if world.playing {
			world.Pause()
		} else {
			world.Play()
		}
		return true, nil

	// speed control
	case ansi.Escape('+'):
		world.playRate *= 2
		world.needsDraw = time.Millisecond
		return true, nil
	case ansi.Escape('-'):
		rate := world.playRate / 2
		if rate <= 0 {
			rate = 1
		}
		if world.playRate != rate {
			world.playRate = rate
		}
		world.needsDraw = time.Millisecond
		return true, nil

	default:
		if hanlded, err := world.View.HandleInput(e, a); hanlded || err != nil {
			return hanlded, err
		}
		return world.World.HandleInput(e, a)
	}
}

// Draw calls Update, renders the world into the screen, and calls DrawPlayOverlay.
func (world *WorldLayer) Draw(screen anansi.Screen, now time.Time) {
	world.init()

	world.Update(now)
	world.View.Draw(screen, now)
	world.DrawPlayOverlay(screen, now)
}

// DrawWorld renders as much of the World as will fit into the screen.
func (world *WorldLayer) DrawWorld(screen anansi.Screen, now time.Time) {
	vp := world.World.Bounds()
	bnd := screen.Bounds()
	if n := vp.Dx() - bnd.Dx(); n > 0 {
		vp.Max.X -= n
	}
	if n := vp.Dy() - bnd.Dy(); n > 0 {
		vp.Max.Y -= n
	}
	world.World.Render(screen.Grid, vp)
}

// DrawPlayOverlay draws an overlay in the upper right to indicate playback state
// and speed.
func (world *WorldLayer) DrawPlayOverlay(screen anansi.Screen, now time.Time) {
	var buf bytes.Buffer
	buf.Grow(128)
	if world.playing {
		buf.WriteRune('▸')
		fmt.Fprintf(&buf, "%v ticks/s", world.playRate)
	} else {
		buf.WriteRune('‖')
	}
	n := MeasureText(buf.Bytes()).Dx()
	bnd := screen.Bounds()
	screen.To(ansi.Pt(bnd.Max.X-n, 1))
	screen.Write(buf.Bytes())
}

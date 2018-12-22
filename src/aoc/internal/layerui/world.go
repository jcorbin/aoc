package layerui

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// World is a discretely timed simulation world driven under a UI.
type World interface {
	// Bounds returns the bounding box of the world space.
	Bounds() image.Rectangle

	// Tick should advance the simulation one step, returning false if the tick
	// wasn't complete/succesful; false return pauses playback.
	Tick() bool

	NeedsDraw() time.Duration
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)

	// Render the world contents to the given grid, starting from the given
	// view offset point.
	Render(g anansi.Grid, viewOffset image.Point)
}

// WorldLayer implements a layer that controls and displays a World
// simulation.
type WorldLayer struct {
	World

	needsDraw time.Duration

	last     time.Time
	ticking  bool
	playing  bool
	playRate int // tick-per-second

	focus      image.Point
	viewOffset image.Point
}

// Play starts playback.
func (world *WorldLayer) Play() {
	world.ticking = true
	world.playing = true
	world.needsDraw = 5 * time.Millisecond
}

// Pause stops playback.
func (world *WorldLayer) Pause() {
	world.ticking = false
	world.playing = false
	world.needsDraw = 5 * time.Millisecond
}

// SetFocus sets the view center point used to determine offset when Draw-ing.
func (world *WorldLayer) SetFocus(p image.Point) {
	world.focus = p
	world.needsDraw = 5 * time.Millisecond
}

// ViewOffset returns the current view offset (as of the last draw).
func (world *WorldLayer) ViewOffset() image.Point {
	return world.viewOffset
}

func (world *WorldLayer) advance(now time.Time) {
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
		for i := 0; i < ticks; i++ {
			if !world.Tick() {
				world.playing = false
				break
			}
		}
		world.last = now
	}

	world.ticking = true
}

// NeedsDraw returns non-zero if the layer needs to be drawn.
func (world *WorldLayer) NeedsDraw() time.Duration {
	nd := world.needsDraw
	if d := world.World.NeedsDraw(); nd == 0 || (d > 0 && d < nd) {
		nd = d
	}
	return nd
}

// HandleInput handles world input: cursor keys to move view focus, '.' to
// step, ' ' to play/pause, and '+'/'-' for playback rate control.
func (world *WorldLayer) HandleInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {
	// arrow keys to move view
	case ansi.CUB, ansi.CUF, ansi.CUU, ansi.CUD:
		if d, ok := ansi.DecodeCursorCardinal(e, a); ok {
			p := world.focus.Add(d)
			bounds := world.Bounds()
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
			if world.focus != p {
				world.focus = p
				world.needsDraw = 5 * time.Millisecond
			}
		}
		return true, nil

	// step
	case ansi.Escape('.'):
		world.needsDraw = 5 * time.Millisecond
		world.ticking = true
		return true, nil

	// play/pause
	case ansi.Escape(' '):
		world.playing = !world.playing
		if world.playing {
			world.last = time.Now()
			if world.playRate == 0 {
				world.playRate = 1
			}
			world.ticking = true
		}
		world.needsDraw = 5 * time.Millisecond
		return true, nil

	// speed control
	case ansi.Escape('+'):
		world.playRate *= 2
		world.needsDraw = 5 * time.Millisecond
		return true, nil
	case ansi.Escape('-'):
		rate := world.playRate / 2
		if rate <= 0 {
			rate = 1
		}
		if world.playRate != rate {
			world.playRate = rate
		}
		world.needsDraw = 5 * time.Millisecond
		return true, nil

	default:
		return world.World.HandleInput(e, a)
	}
}

// Draw advances the world (may call World.Tick as much as needed),
// World.Render()s into the screen grid, and draws a playback control in the
// upper-right corner.
func (world *WorldLayer) Draw(screen anansi.Screen, now time.Time) {
	world.needsDraw = 0
	world.advance(now)
	world.viewOffset = screen.Bounds().Size().Div(2).Sub(world.focus)
	world.Render(screen.Grid, world.viewOffset)
	world.DrawPlayOverlay(screen, now)
	if world.playing {
		world.needsDraw = time.Second / time.Duration(world.playRate)
	}
}

// DrawPlayOverlay draws an overlay in the upper right to indicate playback
// state and speed.
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

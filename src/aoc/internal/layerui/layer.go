package layerui

import (
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
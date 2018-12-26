package anui

import (
	"image"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// BannerLayer provides a header banner overlay.
type BannerLayer struct {
	banner    []byte
	needsDraw time.Duration
}

// Say sets the message string.
func (ban *BannerLayer) Say(mess string) {
	ban.banner = []byte(mess)
	ban.needsDraw = time.Millisecond
}

// HandleInput is a no-op.
func (ban *BannerLayer) HandleInput(e ansi.Escape, a []byte) (bool, error) {
	return false, nil
}

// Draw the banner overlay.
func (ban *BannerLayer) Draw(screen anansi.Screen, now time.Time) {
	ban.needsDraw = 0
	at := screen.Grid.Rect.Min
	bannerWidth := MeasureText(ban.banner).Dx()
	screenWidth := screen.Bounds().Dx()
	at.X += screenWidth/2 - bannerWidth/2
	WriteIntoGrid(screen.Grid.SubAt(at), ban.banner)
}

// NeedsDraw returns non-zero if the layer needs to be drawn (if the mesage
// has changed since last draw).
func (ban *BannerLayer) NeedsDraw() time.Duration {
	return ban.needsDraw
}

// ModalLayer provides an overlay that is displayed screen-centered, and eats
// all input until dismissed.
type ModalLayer struct {
	mess      []byte
	messSize  image.Point
	needsDraw time.Duration
}

// Display sets the messages string.
func (mod *ModalLayer) Display(mess string) {
	if mess == "" {
		mod.mess = nil
		mod.messSize = image.ZP
	} else {
		mod.mess = []byte(mess)
		mod.messSize = MeasureText(mod.mess).Size()
	}
	mod.needsDraw = time.Millisecond
}

// HandleInput eats all input if a message string is set, dismissing the
// message with <Esc>.
func (mod *ModalLayer) HandleInput(e ansi.Escape, a []byte) (bool, error) {
	// no message, ignore
	if mod.mess == nil {
		return false, nil
	}
	switch e {
	// <Esc> to dismiss message
	case ansi.Escape('\x1b'):
		mod.Display("")
		return true, nil
	// eat any other input when a message is shown
	default:
		return true, nil
	}
}

// NeedsDraw returns non-zero if the layer needs to be drawn (if the display
// has been changed since last draw).
func (mod *ModalLayer) NeedsDraw() time.Duration {
	return mod.needsDraw
}

// Draw the modal overlay.
func (mod *ModalLayer) Draw(screen anansi.Screen, now time.Time) {
	mod.needsDraw = 0
	if mod.mess == nil || mod.messSize == image.ZP {
		return
	}
	screenSize := screen.Bounds().Size()
	screenMid := screenSize.Div(2)
	messMid := mod.messSize.Div(2)
	offset := screenMid.Sub(messMid)
	WriteIntoGrid(screen.Grid.SubAt(screen.Grid.Rect.Min.Add(offset)), mod.mess)
}

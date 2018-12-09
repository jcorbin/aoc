package geom

import "image"

// RCore is the core of a rectangular array.
type RCore struct {
	image.Rectangle
	Origin image.Point
	Stride int
}

// Index returns the array index for the given point and true if it's in
// bounds, -1 and false otherwise.
func (rc RCore) Index(p image.Point) (int, bool) {
	if !p.In(rc.Rectangle) {
		return -1, false
	}
	p = p.Sub(rc.Origin)
	return p.Y*rc.Stride + p.X, true
}

// Bounds helps to implement the image.Image interface.
func (rc RCore) Bounds() image.Rectangle { return rc.Rectangle }

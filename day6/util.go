package main

import (
	"image"
	"io"

	"github.com/jcorbin/anansi"
)

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

// TODO haul this back to anansi
func writeGrid(w io.Writer, g anansi.Grid) (int64, error) {
	const border = false

	var buf anansi.Buffer
	var cur anansi.CursorState

	buf.WriteSGR(cur.MergeSGR(0))

	// header
	if border {
		buf.WriteRune('+')
		for x := g.Rect.Min.X; x < g.Rect.Max.X; x++ {
			buf.WriteRune('-')
		}
		buf.WriteRune('+')
		buf.WriteString("\r\n")
		cur.Y++
	}

	if border {
		buf.WriteRune('|')
	}
	for i := 0; i < len(g.Rune); i++ {
		if i > 0 && i%g.Stride == 0 {
			buf.WriteSGR(cur.MergeSGR(0))
			if border {
				buf.WriteString("|\r\n|")
			} else {
				buf.WriteString("\r\n")
			}
			cur.X = 1
			cur.Y++
		}
		gr, ga := g.Rune[i], g.Attr[i]
		if gr == 0 {
			ga = 0
			gr = ' '
		}
		buf.WriteSGR(cur.MergeSGR(ga))
		buf.WriteRune(gr)
		cur.X++
	}

	if cur.X > 1 {
		buf.WriteSGR(cur.MergeSGR(0))
		if border {
			buf.WriteString("|\r\n")
		} else {
			buf.WriteString("\r\n")
		}
		cur.X = 1
		cur.Y++
	}

	// footer
	if border {
		buf.WriteRune('+')
		for x := g.Rect.Min.X; x < g.Rect.Max.X; x++ {
			buf.WriteRune('-')
		}
		buf.WriteRune('+')
		buf.WriteString("\r\n")
		cur.Y++
	}

	return buf.WriteTo(w)
}

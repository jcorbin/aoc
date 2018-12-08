package main

import (
	"image"
	"io"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
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
	var buf anansi.Buffer
	var cur anansi.CursorState
	cur.Point = ansi.Pt(1, 1)
	for y := g.Rect.Min.Y; y < g.Rect.Max.Y; y++ {
		cur = writeGridRow(&buf, cur, g, y)
		buf.WriteString("\r\n")
		cur.X = 1
		cur.Y++
	}
	return buf.WriteTo(w)
}

func writeGridRow(
	buf *anansi.Buffer,
	cur anansi.CursorState,
	g anansi.Grid, row int,
) anansi.CursorState {
	buf.WriteSGR(cur.MergeSGR(0))
	for pt := ansi.Pt(g.Rect.Min.X, row); pt.X < g.Rect.Max.X; pt.X++ {
		i, _ := g.CellOffset(pt)
		gr, ga := g.Rune[i], g.Attr[i]
		if gr == 0 {
			ga = 0
			gr = ' '
		}
		buf.WriteSGR(cur.MergeSGR(ga))
		buf.WriteRune(gr)
		cur.X++
	}
	buf.WriteSGR(cur.MergeSGR(0))
	return cur
}

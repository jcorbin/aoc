package display

import (
	"io"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// WriteGrid just writes the grid's contents to the given io.Writer,
// no cursor positioning or screen clearing is done.
//
// TODO haul this back to anansi
func WriteGrid(w io.Writer, g anansi.Grid) (int64, error) {
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

package worldui

import (
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

func writeIntoGrid(g anansi.Grid, b []byte) ansi.Point {
	var cur anansi.CursorState
	cur.Point = g.Rect.Min
	for len(b) > 0 {
		e, a, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e == 0 {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			e = ansi.Escape(r)
		}
		switch e {
		case ansi.Escape('\n'):
			cur.Y++
			cur.X = g.Rect.Min.X

		case ansi.CSI('m'):
			if attr, _, err := ansi.DecodeSGR(a); err == nil {
				cur.MergeSGR(attr)
			}

		default:
			// write runes into grid, with cursor style, ignoring any other
			// escapes; treating `_` as transparent
			if !e.IsEscape() {
				if i, ok := g.CellOffset(cur.Point); ok {
					if e != ansi.Escape('_') {
						g.Rune[i] = rune(e)
						g.Attr[i] = cur.Attr
					}
				}
				cur.X++
			}
		}
	}
	return cur.Point
}

func measureTextBox(b []byte) (box ansi.Rectangle) {
	box.Min = ansi.Pt(1, 1)
	box.Max = ansi.Pt(1, 1)
	pt := box.Min
	for len(b) > 0 {
		e, _ /*a*/, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e == 0 {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			e = ansi.Escape(r)
		}
		switch e {

		case ansi.Escape('\n'):
			pt.Y++
			pt.X = 1

			// TODO would be nice to borrow cursor movement processing from anansi.Screen et al

		default:
			// ignore escapes, advance on runes
			if !e.IsEscape() {
				pt.X++
			}

		}
		if box.Max.X < pt.X {
			box.Max.X = pt.X
		}
		if box.Max.Y < pt.Y {
			box.Max.Y = pt.Y
		}
	}
	return box
}

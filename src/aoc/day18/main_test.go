package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_board_step(t *testing.T) {
	// Initial state:
	brd := testBoard(
		".#.#...|#.",
		".....#|##|",
		".|..|...#.",
		"..|#.....#",
		"#.#|||#|#|",
		"...#.||...",
		".|....|...",
		"||...#|.#|",
		"|.||||..|.",
		"...#.|..|.",
	)

	for i, next := range []board{
		// After 1 minute:
		testBoard(
			".......##.",
			"......|###",
			".|..|...#.",
			"..|#||...#",
			"..##||.|#|",
			"...#||||..",
			"||...|||..",
			"|||||.||.|",
			"||||||||||",
			"....||..|.",
		),

		// After 2 minutes:
		testBoard(
			".......#..",
			"......|#..",
			".|.|||....",
			"..##|||..#",
			"..###|||#|",
			"...#|||||.",
			"|||||||||.",
			"||||||||||",
			"||||||||||",
			".|||||||||",
		),

		// After 3 minutes:
		testBoard(
			".......#..",
			"....|||#..",
			".|.||||...",
			"..###|||.#",
			"...##|||#|",
			".||##|||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
		),

		// After 4 minutes:
		testBoard(
			".....|.#..",
			"...||||#..",
			".|.#||||..",
			"..###||||#",
			"...###||#|",
			"|||##|||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
		),

		// After 5 minutes:
		testBoard(
			"....|||#..",
			"...||||#..",
			".|.##||||.",
			"..####|||#",
			".|.###||#|",
			"|||###||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
		),

		// After 6 minutes:
		testBoard(
			"...||||#..",
			"...||||#..",
			".|.###|||.",
			"..#.##|||#",
			"|||#.##|#|",
			"|||###||||",
			"||||#|||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
		),

		// After 7 minutes:
		testBoard(
			"...||||#..",
			"..||#|##..",
			".|.####||.",
			"||#..##||#",
			"||##.##|#|",
			"|||####|||",
			"|||###||||",
			"||||||||||",
			"||||||||||",
			"||||||||||",
		),

		// After 8 minutes:
		testBoard(
			"..||||##..",
			"..|#####..",
			"|||#####|.",
			"||#...##|#",
			"||##..###|",
			"||##.###||",
			"|||####|||",
			"||||#|||||",
			"||||||||||",
			"||||||||||",
		),

		// After 9 minutes:
		testBoard(
			"..||###...",
			".||#####..",
			"||##...##.",
			"||#....###",
			"|##....##|",
			"||##..###|",
			"||######||",
			"|||###||||",
			"||||||||||",
			"||||||||||",
		),

		// After 10 minutes:
		testBoard(
			".||##.....",
			"||###.....",
			"||##......",
			"|##.....##",
			"|##.....##",
			"|##....##|",
			"||##.####|",
			"||#####|||",
			"||||#|||||",
			"||||||||||",
		),
	} {
		tmp := make([]byte, len(brd.d))
		brd.step(tmp)
		brd.d = tmp
		assert.Equal(t,
			boardLines(next),
			boardLines(brd),
			"expected board[%v]", 1+i)
	}
}

func boardLines(brd board) []string {
	lines := make([]string, 0, brd.Stride)
	var buf bytes.Buffer
	for i := 0; i < len(brd.d); {
		buf.Reset()
		for j := 0; j < brd.Stride; j++ {
			buf.WriteByte(brd.d[i])
			i++
		}
		lines = append(lines, buf.String())
	}
	return lines
}

func testBoard(lines ...string) (brd board) {
	n := len(lines)
	brd.d = make([]byte, 0, n*n)
	for _, line := range lines {
		if len(line) != n {
			panic("board lines not square")
		}
		for i := 0; i < len(line); i++ {
			brd.d = append(brd.d, line[i])
		}
	}
	brd.Stride = n
	brd.Max.X = n
	brd.Max.Y = n
	return brd
}

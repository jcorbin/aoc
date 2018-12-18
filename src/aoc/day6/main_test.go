package main

import (
	"bytes"
	"image"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_problem_solve(t *testing.T) {
	tc := struct {
		in        []image.Point
		gridLines []string
		counts    map[int]int
	}{
		in: []image.Point{
			image.Pt(1, 1), // #1 A
			image.Pt(1, 6), // #2 B
			image.Pt(8, 3), // #3 C
			image.Pt(3, 4), // #4 D
			image.Pt(5, 5), // #5 E
			image.Pt(8, 9), // #6 F
		},
		gridLines: []string{

			// "aaaaa.cccc",
			// "aAaaa.cccc",
			// "aaaddecccc",
			// "aadddeccCc",
			// "..dDdeeccc",
			// "bb.deEeecc",
			// "bBb.eeee..",
			// "bbb.eeefff",
			// "bbb.eeffff",
			// "bbb.ffffFf",

			"Aaaa.ccc",
			"aaddeccc",
			"adddeccC",
			".dDdeecc",
			"b.deEeec",
			"Bb.eeee.",
			"bb.eeeff",
			"bb.eefff",
			"bb.ffffF",
		},
		counts: map[int]int{
			4: 9,  // D
			5: 17, // E
		},
	}

	var prob ui
	prob.points = tc.in
	prob.init()
	require.NoError(t, prob.populate())
	prob.render()

	var buf bytes.Buffer
	for i, r := range prob.g.Rune {
		if i > 0 && i%prob.g.Stride == 0 {
			buf.WriteRune('\n')
		}
		buf.WriteRune(r)
	}
	assert.Equal(t, tc.gridLines, strings.Split(buf.String(), "\n"))
	assert.Equal(t, tc.counts, prob.countArea())
}

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_board_search(t *testing.T) {
	for _, tc := range []struct {
		pattern []uint8
		after   int
	}{
		{[]uint8{5, 1, 5, 8, 9}, 9},
		{[]uint8{0, 1, 2, 4, 5}, 5},
		{[]uint8{9, 2, 5, 1, 0}, 18},
		{[]uint8{5, 9, 4, 1, 4}, 2018},
		{[]uint8{3, 5, 0, 4, 1}, 1215551},
	} {
		t.Run(fmt.Sprintf("%v", tc.pattern), func(t *testing.T) {
			var brd board
			brd.init(0, 1, 3, 7)
			assert.Equal(t, tc.after, brd.search(tc.pattern))
		})
	}
}

func Benchmark_board_search(b *testing.B) {
	for _, bc := range []struct {
		pattern []uint8
		after   int
	}{
		{[]uint8{5, 1, 5, 8, 9}, 9},
		{[]uint8{0, 1, 2, 4, 5}, 5},
		{[]uint8{9, 2, 5, 1, 0}, 18},
		{[]uint8{5, 9, 4, 1, 4}, 2018},
		{[]uint8{3, 5, 0, 4, 1}, 1215551},
	} {
		b.Run(fmt.Sprintf("%v", bc.pattern), func(b *testing.B) {
			var brd board
			brd.scores = make([]uint8, 0, 2*bc.after)
			for i := 0; i < b.N; i++ {
				brd.scores = brd.scores[:0]
				brd.init(0, 1, 3, 7)
				_ = brd.search(bc.pattern)
			}
		})
	}
}

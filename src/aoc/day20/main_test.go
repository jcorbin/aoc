package main

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCases = []struct {
	pattern  string
	furthest int
	lines    []string
}{
	{`^WNE$`, 3, []string{
		"#####",
		"#.|.#",
		"#-###",
		"#.|X#",
		"#####",
	}},

	// {`^ENWWW$`, []string{
	// 	"#########",
	// 	"#.|.|.|.#",
	// 	"#######-#",
	// 	"    #X|.#",
	// 	"    #####",
	// }},

	// {`^ENWWWNEEE$`, []string{
	// 	"#########",
	// 	"#.|.|.|.#",
	// 	"#-#######",
	// 	"#.|.|.|.#",
	// 	"#######-#",
	// 	"    #X|.#",
	// 	"    #####",
	// }},

	// {`^ENWWW(NEEE|SSE)$`, []string{
	// 	"#########",
	// 	"#.|.|.|.#",
	// 	"#-#######",
	// 	"#.|.|.|.#",
	// 	"#-#####-#",
	// 	"#.# #X|.#",
	// 	"#-#######",
	// 	"#.|.#    ",
	// 	"#####    ",
	// }},

	{`^ENWWW(NEEE|SSE(EE|N))$`, 10, []string{
		"#########",
		"#.|.|.|.#",
		"#-#######",
		"#.|.|.|.#",
		"#-#####-#",
		"#.#.#X|.#",
		"#-#-#####",
		"#.|.|.|.#",
		"#########",
	}},

	{`^ENNWSWW(NEWS|)SSSEEN(WNSE|)EE(SWEN|)NNN$`, 18, []string{
		"###########",
		"#.|.#.|.#.#",
		"#-###-#-#-#",
		"#.|.|.#.#.#",
		"#-#####-#-#",
		"#.#.#X|.#.#",
		"#-#-#####-#",
		"#.#.|.|.|.#",
		"#-###-###-#",
		"#.|.|.#.|.#",
		"###########",
	}},

	{`^ESSWWN(E|NNENN(EESS(WNSE|)SSS|WWWSSSSE(SW|NNNE)))$`, 23, []string{
		"#############",
		"#.|.|.|.|.|.#",
		"#-#####-###-#",
		"#.#.|.#.#.#.#",
		"#-#-###-#-#-#",
		"#.#.#.|.#.|.#",
		"#-#-#-#####-#",
		"#.#.#.#X|.#.#",
		"#-#-#-###-#-#",
		"#.|.#.|.#.#.#",
		"###-#-###-#-#",
		"#.|.#.|.|.#.#",
		"#############",
	}},

	{`^WSSEESWWWNW(S|NENNEEEENN(ESSSSW(NWSW|SSEN)|WSWWN(E|WWS(E|SS))))$`, 31, []string{
		"###############",
		"#.|.|.|.#.|.|.#",
		"#-###-###-#-#-#",
		"#.|.#.|.|.#.#.#",
		"#-#########-#-#",
		"#.#.|.|.|.|.#.#",
		"#-#-#########-#",
		"#.#.#.|X#.|.#.#",
		"###-#-###-#-#-#",
		"#.|.#.#.|.#.|.#",
		"#-###-#####-###",
		"#.|.#.|.|.#.#.#",
		"#-#-#####-#-#-#",
		"#.#.|.|.|.#.|.#",
		"###############",
	}},
}

// func Test_build_maps(t *testing.T) {
// 	for _, tc := range testCases {
// 		t.Run(tc.pattern, func(t *testing.T) {
// 			var bld builder
// 			err := bld.buildRooms(tc.pattern)
// 			require.NoError(t, err)
// 			var rm roomMap
// 			bld.pg.draw(&rm, image.ZP)
// 			t.Logf("map bounds %v", rm.bounds)
// 			assert.Equal(t,
// 				tc.lines,
// 				strings.Split(rm.draw(), "\n"))
// 		})
// 	}
// }

func Test_explore_rooms(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.pattern, func(t *testing.T) {
			var bld builder
			err := bld.buildRooms(tc.pattern)
			require.NoError(t, err)
			assert.Equal(t,
				tc.furthest,
				bld.pg.fill(image.ZP, 0, make(pointScore, len(bld.pg))),
			)
		})
	}
}

func Benchmark_buildMap(b *testing.B) {
	for _, bc := range testCases {
		b.Run(bc.pattern, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var bld builder
				_ = bld.buildRooms(bc.pattern)
			}
		})
	}
}

package main

import (
	"bufio"
	"bytes"
	"image"
	"log"
	"testing"

	"github.com/jcorbin/anansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetFlags(0)
}

func Test_gameWorld_combat(t *testing.T) {
	type result struct {
		rounds   int
		team     string
		remainHP int
	}

	type checkpoint struct {
		round int
		lines []string
	}

	type scenario struct {
		name        string
		initialGrid []string
		checkpoints []checkpoint
		result
	}

	for _, tc := range []scenario{

		{
			name: "ex1",
			initialGrid: []string{
				"#######",
				"#.G...#", // G(200)
				"#...EG#", // E(200) G(200)
				"#.#.#G#", // G(200)
				"#..G#E#", // G(200) E(200)
				"#.....#",
				"#######",
			},
			checkpoints: []checkpoint{
				{1, []string{
					"#######",
					"#..G..# G(200)",
					"#...EG# E(197) G(197)",
					"#.#G#G# G(200) G(197)",
					"#...#E# E(197)",
					"#.....#",
					"#######",
				}},

				{2, []string{
					"#######",
					"#...G.# G(200)",
					"#..GEG# G(200) E(188) G(194)",
					"#.#.#G# G(194)",
					"#...#E# E(194)",
					"#.....#",
					"#######",
				}},

				{22, []string{
					"#######",
					"#...G.# G(200)",
					"#..GEG# G(200) E(8) G(134)",
					"#.#.#G# G(134)",
					"#...#E# E(134)",
					"#.....#",
					"#######",
				}},

				// attack G@(4,1)#10 => E@(4,2)#20
				// attack G@(3,2)#36 => E@(4,2)#20
				// attack E@(4,2)#20 => G@(5,2)#22
				// attack G@(5,2)#22 => E@(4,2)#20
				// attack E@(5,4)#39 => G@(5,3)#30
				// attack E@(5,4)#39 => G@(5,3)#30
				// finished round {23 none 793} (1 elves vs 4 goblins)

				// Combat ensues; eventually, the top Elf dies:
				{23, []string{
					"#######",
					"#...G.# G(200)",
					"#..G.G# G(200) G(131)",
					"#.#.#G# G(131)",
					"#...#E# E(131)",
					"#.....#",
					"#######",
				}},

				{24, []string{
					"#######",
					"#..G..# G(200)",
					"#...G.# G(131)",
					"#.#G#G# G(200) G(128)",
					"#...#E# E(128)",
					"#.....#",
					"#######",
				}},

				{25, []string{
					"#######",
					"#.G...# G(200)",
					"#..G..# G(131)",
					"#.#.#G# G(125)",
					"#..G#E# G(200) E(125)",
					"#.....#",
					"#######",
				}},

				{26, []string{
					"#######",
					"#G....# G(200)",
					"#.G...# G(131)",
					"#.#.#G# G(122)",
					"#...#E# E(122)",
					"#..G..# G(200)",
					"#######",
				}},

				{27, []string{
					"#######",
					"#G....# G(200)",
					"#.G...# G(131)",
					"#.#.#G# G(119)",
					"#...#E# E(119)",
					"#...G.# G(200)",
					"#######",
				}},

				{28, []string{
					"#######",
					"#G....# G(200)",
					"#.G...# G(131)",
					"#.#.#G# G(116)",
					"#...#E# E(113)",
					"#....G# G(200)",
					"#######",
				}},

				// More combat ensues; eventually, the bottom Elf dies:
				{47, []string{
					"#######",
					"#G....# G(200)",
					"#.G...# G(131)",
					"#.#.#G# G(59)",
					"#...#.#",
					"#....G# G(200)",
					"#######",
				}},
			},
			result: result{
				rounds:   47,
				team:     "goblins",
				remainHP: 590,
			},
		},

		{
			name: "ex2",
			initialGrid: []string{
				"#######",
				"#G..#E#",
				"#E#E.E#",
				"#G.##.#",
				"#...#E#",
				"#...E.#",
				"#######",
			},
			checkpoints: []checkpoint{
				{47, []string{
					"#######",
					"#...#E# E(200)",
					"#E#...# E(197)",
					"#.E##.# E(185)",
					"#E..#E# E(200) E(200)",
					"#.....#",
					"#######",
				}},
			},
			result: result{
				rounds:   37,
				team:     "elves",
				remainHP: 982,
			},
		},

		/*

			#######       #######
			#E..EG#       #.E.E.#   E(164), E(197)
			#.#G.E#       #.#E..#   E(200)
			#E.##E#  -->  #E.##.#   E(98)
			#G..#.#       #.E.#.#   E(200)
			#..E#.#       #...#.#
			#######       #######

			Combat ends after 46 full rounds
			Elves win with 859 total hit points left
			Outcome: 46 * 859 = 39514

			#######       #######
			#E.G#.#       #G.G#.#   G(200), G(98)
			#.#G..#       #.#G..#   G(200)
			#G.#.G#  -->  #..#..#
			#G..#.#       #...#G#   G(95)
			#...E.#       #...G.#   G(200)
			#######       #######

			Combat ends after 35 full rounds
			Goblins win with 793 total hit points left
			Outcome: 35 * 793 = 27755

			#######       #######
			#.E...#       #.....#
			#.#..G#       #.#G..#   G(200)
			#.###.#  -->  #.###.#
			#E#G#G#       #.#.#.#
			#...#G#       #G.G#G#   G(98), G(38), G(200)
			#######       #######

			Combat ends after 54 full rounds
			Goblins win with 536 total hit points left
			Outcome: 54 * 536 = 28944

			#########       #########
			#G......#       #.G.....#   G(137)
			#.E.#...#       #G.G#...#   G(200), G(200)
			#..##..G#       #.G##...#   G(200)
			#...##..#  -->  #...##..#
			#...#...#       #.G.#...#   G(200)
			#.G...G.#       #.......#
			#.....G.#       #.......#
			#########       #########

			Combat ends after 20 full rounds
			Goblins win with 937 total hit points left
			Outcome: 20 * 937 = 18740

		*/
	} {

		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			for _, line := range tc.initialGrid {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}

			var world gameWorld
			world.load(&buf)

			res := func() result {
				hp := world.remainingHP()
				return result{
					rounds:   world.round,
					team:     world.winningTeam(),
					remainHP: hp,
				}
			}

			var g anansi.Grid

			ci := 0
			for i := 0; i < 10*tc.rounds && world.tick(); i++ {
				sc := bufio.NewScanner(&logs)
				for sc.Scan() {
					t.Logf("%s\n", sc.Bytes())
				}
				t.Logf("finished round %v (%v elves vs %v goblins)\n", res(), len(world.elves), len(world.goblins))

				g.Resize(world.bounds.Size().Add(image.Pt(100, 1)))
				for i := range g.Rune {
					g.Rune[i] = 0
					g.Attr[i] = 0
				}
				world.render(g, image.ZP)
				lines := gridLines(g)

				logGrid := true
				if ci < len(tc.checkpoints) {
					chk := tc.checkpoints[ci]
					require.False(t, chk.round < world.round, "missed checkpoint[%v]", chk.round)
					if chk.round == world.round {
						require.Equal(t, chk.lines, lines, "expected checkpoint[%v]", chk.round)
						t.Logf("passed checpoint[%v]", chk.round)
						logGrid = false
						ci++
					}
				}

				if logGrid {
					ruler := ""
					for i, x := 0, world.bounds.Dx(); i <= x; i++ {
						ruler += string('0' + i%10)
					}
					t.Logf("   %s", ruler)
					for i, line := range lines {
						t.Logf("%d: %s", i, line)
					}
				}

			}
			assert.Equal(t, tc.result, res())
		})
	}
}

func gridLines(g anansi.Grid) []string {
	var buf bytes.Buffer
	buf.Grow(g.Stride)
	lines := make([]string, 0, g.Rect.Dy())
	for pt := g.Rect.Min; pt.Y < g.Rect.Max.Y; pt.Y++ {
		buf.Reset()
		nz := 0
		for pt.X = g.Rect.Min.X; pt.X < g.Rect.Max.X; pt.X++ {
			i, _ := g.CellOffset(pt)
			r := g.Rune[i]
			if r == 0 {
				nz++
			} else {
				for i := 0; i < nz; i++ {
					buf.WriteRune(' ')
				}
				nz = 0
				buf.WriteRune(r)
			}
		}
		lines = append(lines, buf.String())
	}
	return lines
}

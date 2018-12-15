package main

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_gameWorld_outcome(t *testing.T) {
	type result struct {
		rounds   int
		team     string
		remainHP int
	}

	type scenario struct {
		name        string
		initialGrid []string
		finalGrid   []string
		result
	}

	for _, tc := range []scenario{

		{
			name: "ex1",
			initialGrid: []string{
				"#######",
				"#G..#E#",
				"#E#E.E#",
				"#G.##.#",
				"#...#E#",
				"#...E.#",
				"#######",
			},
			finalGrid: []string{
				"#######",
				"#...#E#", // E(200)
				"#E#...#", // E(197)
				"#.E##.#", // E(185)
				"#E..#E#", // E(200) E(200)
				"#.....#",
				"#######",
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

			for i := 0; i < 10*tc.rounds && world.tick(); i++ {
				sc := bufio.NewScanner(&logs)
				for sc.Scan() {
					fmt.Printf("+ %s\n", sc.Bytes())
				}
				fmt.Printf("finished round %v\n", res())
				fmt.Printf("WUT %v v %v\n", len(world.elves), len(world.goblins))
			}
			assert.Equal(t, tc.result, res())

			// TODO finalGrid check

		})
	}
}

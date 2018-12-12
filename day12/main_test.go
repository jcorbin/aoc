package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testSpace space

func init() {
	var err error
	testSpace, err = read(bytes.NewReader([]byte(
		`initial state: #..#.#..##......###...###

...## => #
..#.. => #
.#... => #
.#.#. => #
.#.## => #
.##.. => #
.#### => #
#.#.# => #
#.### => #
##.#. => #
##.## => #
###.. => #
###.# => #
####. => #
`)))
	if err != nil {
		panic(err.Error())
	}
}

func testSpaceCopy() space {
	return space{
		rules:  testSpace.rules,
		offset: testSpace.offset,
		chunks: append([]uint64(nil), testSpace.chunks...),
	}
}

func Test_space_tick(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rules map[ruleKey]bool
		init  uint64
		res   []uint64
	}{

		{
			name: "prior l2",
			rules: map[ruleKey]bool{
				rule(false, false, false, false, true): true,
			},
			init: 1,
			res:  []uint64{1 << 62, 0},
		},

		{
			name: "prior l1",
			rules: map[ruleKey]bool{
				rule(false, false, false, true, false): true,
			},
			init: 1,
			res:  []uint64{1 << 63, 0},
		},

		{
			name: "next r1",
			rules: map[ruleKey]bool{
				rule(false, true, false, false, false): true,
			},
			init: 1 << 63,
			res:  []uint64{0, 1},
		},

		{
			name: "next r2",
			rules: map[ruleKey]bool{
				rule(true, false, false, false, false): true,
			},
			init: 1 << 63,
			res:  []uint64{0, 2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var spc space
			spc.chunks = []uint64{tc.init}
			for rk, v := range tc.rules {
				spc.rules[rk] = v
			}
			spc.tick()
			if !assert.Equal(t, tc.res, spc.chunks) {
				t.Logf("@%v %b", spc.offset, spc.chunks)
			}
		})
	}
}

func Test_space_sim(t *testing.T) {
	spc := testSpaceCopy()
	for i := 0; i < 20; i++ {
		spc.tick()
	}
	var buf bytes.Buffer
	spc.writePotBytes(&buf)
	assert.Equal(t, -2, spc.min())
	assert.Equal(t, 35, spc.max())
	assert.Equal(t, "#....##....#####...#######....#.#..##", buf.String())
	assert.Equal(t, 325, spc.sumPots())
}

func Benchmark_space_tick(b *testing.B) {
	spc := testSpaceCopy()
	for i := 0; i < b.N; i++ {
		spc.tick()
	}
}

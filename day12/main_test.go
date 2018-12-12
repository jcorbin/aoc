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
	spc := space{
		rules: testSpace.rules,
		min:   testSpace.min,
		max:   testSpace.max,
	}
	spc.pots = make(map[int]bool, len(testSpace.pots))
	for i, v := range testSpace.pots {
		spc.pots[i] = v
	}
	return spc
}

func Test_space_tick(t *testing.T) {
	spc := testSpaceCopy()
	for i := 0; i < 20; i++ {
		spc.tick()
	}
	t.Logf("wut %#v", spc)
	var buf bytes.Buffer
	spc.writePotBytes(&buf)
	assert.Equal(t, -2, spc.min)
	assert.Equal(t, 35, spc.max)
	assert.Equal(t, "#....##....#####...#######....#.#..##", buf.String())
	assert.Equal(t, 325, spc.sumPots())
}

func Benchmark_space_tick(b *testing.B) {
	spc := testSpaceCopy()
	for i := 0; i < b.N; i++ {
		spc.tick()
	}
}

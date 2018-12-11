package main

import (
	"fmt"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_powerLevel(t *testing.T) {
	for _, tc := range []struct {
		pt     image.Point
		serial int
		lvl    int
	}{
		{image.Pt(3, 5), 8, 4},
		{image.Pt(122, 79), 57, -5},
		{image.Pt(217, 196), 39, 0},
		{image.Pt(101, 153), 71, 4},
	} {
		t.Run(fmt.Sprintf("@%v serial:%v", tc.pt, tc.serial), func(t *testing.T) {
			assert.Equal(t, tc.lvl, powerLevel(tc.pt, tc.serial))
		})
	}
}

var testSolutions = []struct {
	name    string
	factory func(serial int, bounds image.Rectangle) solver
}{
	{"fuelGrid", fuelGridSolver},
	{"spec", specSolver},
}

func Test_solvers(t *testing.T) {
	type result struct {
		loc   image.Point
		level int
	}

	testCases := []struct {
		serial int
		size   int
		result
	}{
		{18, 3, result{image.Pt(33, 45), 29}},
		{42, 3, result{image.Pt(21, 61), 30}},
		{18, 16, result{image.Pt(90, 269), 113}},
		{42, 12, result{image.Pt(232, 251), 119}},
	}

	bounds := image.Rect(1, 1, 301, 301)

	for _, sol := range testSolutions {
		t.Run(sol.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("serial:%d size:%v", tc.serial, tc.size), func(t *testing.T) {
					solver := sol.factory(tc.serial, bounds)
					loc, level := solver.solve(tc.size)
					assert.Equal(t, tc.result, result{loc, level})
				})
			}
		})
	}
}

func Benchmark_solvers(b *testing.B) {
	benchCases := []struct {
		serial int
		size   int
	}{
		{18, 3},
		{18, 16},

		{42, 1},
		{42, 2},
		{42, 3},
		{42, 4},
		{42, 5},
		{42, 6},
		{42, 7},
		{42, 8},
		{42, 9},
		{42, 10},
		{42, 11},
		{42, 12},
		{42, 13},
		{42, 14},
		{42, 15},
		{42, 16},
		{42, 17},
		{42, 18},
		{42, 19},
		{42, 20},
	}

	bounds := image.Rect(1, 1, 301, 301)

	for _, sol := range testSolutions {
		b.Run(sol.name, func(b *testing.B) {
			for _, bc := range benchCases {
				b.Run(fmt.Sprintf("serial:%d size:%d", bc.serial, bc.size), func(b *testing.B) {
					solver := sol.factory(bc.serial, bounds)
					for i := 0; i < b.N; i++ {
						_, _ = solver.solve(bc.size)
					}
				})
			}
		})
	}
}

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

func Test_solvers(t *testing.T) {
	solutions := []struct {
		name    string
		factory func(serial int, bounds image.Rectangle) solver
	}{
		{"fuelGrid", fuelGridSolver},
		{"spec", specSolver},
	}

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

	for _, sol := range solutions {
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

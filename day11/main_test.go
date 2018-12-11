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

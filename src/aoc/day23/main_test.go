package main

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_point3_eachWithin(t *testing.T) {
	for _, tc := range []struct {
		p point3
		r int
		e []point3
	}{
		{point3{0, 0, 0}, 0, []point3{
			{0, 0, 0},
		}},
		{point3{0, 0, 0}, 1, []point3{
			{0, 0, -1},
			{0, -1, 0},
			{-1, 0, 0},
			{0, 0, 0},
			{1, 0, 0},
			{0, 1, 0},
			{0, 0, 1},
		}},
	} {
		t.Run(fmt.Sprintf("p:%v r:%v", tc.p, tc.r), func(t *testing.T) {
			seen := make(map[point3]struct{}, tc.r*tc.r*tc.r)
			tc.p.eachWithin(tc.r, func(p point3) {
				seen[p] = struct{}{}
			})
			r := make([]point3, 0, len(seen))
			for p := range seen {
				r = append(r, p)
			}
			sort.Slice(r, func(i, j int) bool {
				if r[i].Z < r[j].Z {
					return true
				}
				if r[i].Z > r[j].Z {
					return false
				}
				if r[i].Y < r[j].Y {
					return true
				}
				if r[i].Y > r[j].Y {
					return false
				}
				return r[i].X < r[j].X
			})
			assert.Equal(t, tc.e, r)
		})
	}
}

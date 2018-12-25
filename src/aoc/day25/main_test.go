package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_clusterPoints(t *testing.T) {
	for i, tc := range []struct {
		points    []point4
		r         int
		nclusters int
	}{
		// 1
		{
			points: []point4{
				{0, 0, 0, 0},
				{3, 0, 0, 0},
				{0, 3, 0, 0},
				{0, 0, 3, 0},
				{0, 0, 0, 3},
				{0, 0, 0, 6},
				{9, 0, 0, 0},
				{12, 0, 0, 0},
			},
			r:         3,
			nclusters: 2,
		},

		// 2
		{
			points: []point4{
				{-1, 2, 2, 0},
				{0, 0, 2, -2},
				{0, 0, 0, -2},
				{-1, 2, 0, 0},
				{-2, -2, -2, 2},
				{3, 0, 2, -1},
				{-1, 3, 2, 2},
				{-1, 0, -1, 0},
				{0, 2, 1, -2},
				{3, 0, 0, 0},
			},
			r:         3,
			nclusters: 4,
		},

		// 3
		{
			points: []point4{
				{1, -1, 0, 1},
				{2, 0, -1, 0},
				{3, 2, -1, 0},
				{0, 0, 3, 1},
				{0, 0, -1, -1},
				{2, 3, -2, 0},
				{-2, 2, 0, 0},
				{2, -2, 0, -1},
				{1, -1, 0, -1},
				{3, 2, 0, 2},
			},
			r:         3,
			nclusters: 3,
		},

		// 4
		{
			points: []point4{
				{1, -1, -1, -2},
				{-2, -2, 0, 1},
				{0, 2, 1, 3},
				{-2, 3, -2, 1},
				{0, 2, 3, -2},
				{-1, -1, 1, -2},
				{0, -2, -1, 0},
				{-2, 2, 3, -1},
				{1, 2, 2, 0},
				{-1, -2, 0, -2},
			},
			r:         3,
			nclusters: 8,
		},
	} {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			if clusters := clusterPoints(tc.points, tc.r); !assert.Equal(t, tc.nclusters, len(clusters)) {
				for j, cluster := range clusters {
					t.Logf("cluster[%v]: %v", j, cluster)
				}
			}
		})
	}
}

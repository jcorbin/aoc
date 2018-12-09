package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_game_run(t *testing.T) {
	for _, tc := range []struct {
		nPlayers   int
		lastMarble int
		highScore  int
	}{
		{9, 25, 32},
		{10, 1618, 8317},
		{13, 7999, 146373},
		{17, 1104, 2764},
		{21, 6111, 54718},
		{30, 5807, 37305},
	} {
		t.Run(fmt.Sprintf("n:%v m:%v", tc.nPlayers, tc.lastMarble), func(t *testing.T) {
			var g game
			g.run(tc.nPlayers, tc.lastMarble)
			_, best := g.highestScore()
			assert.Equal(t, tc.highScore, best)
		})
	}
}

func Benchmark_game_run(b *testing.B) {
	for _, bc := range []struct {
		nPlayers   int
		lastMarble int
	}{
		{9, 25},
		{10, 1618},
		{13, 7999},
		{17, 1104},
		{21, 6111},
		{30, 5807},
		{100, 100},
		{100, 200},
		{100, 400},
		{100, 800},
		{100, 1600},
		{100, 3200},
		{100, 6400},
		{100, 12800},
		{100, 25600},
		{100, 51200},
	} {
		b.Run(fmt.Sprintf("n:%v m:%v", bc.nPlayers, bc.lastMarble), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var g game
				g.run(bc.nPlayers, bc.lastMarble)
			}
		})
	}
}

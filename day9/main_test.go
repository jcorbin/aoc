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

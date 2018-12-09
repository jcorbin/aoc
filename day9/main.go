package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jcorbin/anansi"
)

var (
	verbose  = flag.Bool("v", false, "verbose")
	nPlayers = flag.Int("n", 9, "number of players")
	mValue   = flag.Int("m", 25, "highest marble")
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

func run(in, out *os.File) error {
	var g game
	g.run(*nPlayers, *mValue)
	besti, best := g.highestScore()
	log.Printf("best was %v by player %v", best, besti+1)
	return nil
}

type game struct {
	playeri int
	scores  []int // for each player

	marblei int
	marbles []int
}

func (g *game) init(n, m int) {
	g.scores = make([]int, n)
	g.marbles = make([]int, 1, m)
}

func (g *game) run(nPlayers, lastMarble int) {
	g.init(nPlayers, lastMarble+1)
	if *verbose {
		log.Printf("%v", g)
	}
	for v := 1; v < lastMarble+1; v++ {
		g.place(v)
		if *verbose {
			log.Printf("v:%v %v", v, g)
		}
	}
}

func (g *game) highestScore() (besti, best int) {
	for i, score := range g.scores {
		if score > best {
			besti, best = i, score
		}
	}
	return besti, best
}

func (g *game) place(v int) {
	if v%23 == 0 {
		i := (g.marblei - 7 + len(g.marbles)) % len(g.marbles)
		t := g.marbles[i]
		copy(g.marbles[i:], g.marbles[i+1:])
		g.marbles = g.marbles[:len(g.marbles)-1]
		g.scores[g.playeri] += v + t
		g.marblei = i
	} else {
		i := g.marblei + 1
		j := i + 1
		im := i % len(g.marbles)
		jm := j % len(g.marbles)
		if im == len(g.marbles)-1 /*im == jm*/ {
			g.marblei = len(g.marbles)
			g.marbles = append(g.marbles, v)
		} else {
			g.marbles = append(g.marbles, 0)
			copy(g.marbles[jm+1:], g.marbles[jm:])
			g.marbles[jm] = v
			g.marblei = jm
		}
	}
	g.playeri = (g.playeri + 1) % len(g.scores)
}

func (g game) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[%d]", g.playeri)
	for i, v := range g.marbles {
		if i == g.marblei {
			fmt.Fprintf(&buf, " (%d)", v)
		} else {
			fmt.Fprintf(&buf, " %d", v)
		}
	}
	return buf.String()
}

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

type el struct {
	p, n *el
	v    int
}

type game struct {
	playeri int
	scores  []int // for each player

	marbles *el
	curm    *el
}

func (g *game) init(n, m int) {
	g.scores = make([]int, n)
	e := &el{v: 0}
	g.marbles = e
	g.curm = g.marbles
	e.n = e
	e.p = e
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
		for i := 0; i < 7; i++ {
			g.curm = g.curm.p
		}
		e := g.curm
		t := e.v
		e.p.n, e.n.p = e.n, e.p
		g.curm = e.n
		if g.marbles == e {
			g.marbles = g.curm
		}
		g.scores[g.playeri] += v + t
	} else {
		g.curm = g.curm.n
		e := &el{v: v}
		e.p = g.curm
		e.n = g.curm.n
		g.curm.n.p = e
		g.curm.n = e
		g.curm = e
	}
	g.playeri = (g.playeri + 1) % len(g.scores)
}

func (g game) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[%d]", g.playeri)
	e := g.marbles
	s := e
	for {
		if e == g.curm {
			fmt.Fprintf(&buf, " (%d)", e.v)
		} else {
			fmt.Fprintf(&buf, " %d", e.v)
		}
		e = e.n
		if e == s {
			break
		}
	}
	return buf.String()
}

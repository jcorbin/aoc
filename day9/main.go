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

func newEl(v int, p, n *el) *el {
	e := &el{v: v}
	if n == nil {
		e.n = e
	} else {
		e.n = n
	}
	if p == nil {
		e.p = e
	} else {
		e.p = p
	}
	return e
}

func (e *el) prev() *el { return e.p }
func (e *el) next() *el { return e.n }

func (e *el) prevN(n int) *el {
	for i := 0; i < n; i++ {
		e = e.p
	}
	return e
}
func (e *el) nextN(n int) *el {
	for i := 0; i < n; i++ {
		e = e.n
	}
	return e
}

func (e *el) unlink() *el {
	e.p.n, e.n.p = e.n, e.p
	r := e.n
	e.p, e.n = nil, nil
	return r
}

func (e *el) append(v int) *el {
	r := newEl(v, e, e.n)
	e.n.p = r
	e.n = r
	return r
}

type game struct {
	playeri int
	scores  []int // for each player

	marbles *el
	curm    *el
}

func (g *game) init(n, m int) {
	g.scores = make([]int, n)
	g.marbles = newEl(0, nil, nil)
	g.curm = g.marbles
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
	switch {
	case v%23 == 0:
		e := g.curm.prevN(7)
		t := e.v
		g.curm = e.unlink()
		if g.marbles == e {
			g.marbles = g.curm
		}
		g.scores[g.playeri] += v + t
	default:
		g.curm = g.curm.next().append(v)
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

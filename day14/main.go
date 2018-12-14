package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type board struct {
	e1, e2 int
	scores []uint8
}

var (
	after   = flag.Int("after", 0, "after N rounds")
	search  = flag.String("search", "", "search for N pattern")
	verbose = flag.Bool("v", false, "verbose")
)

func run(in, out *os.File) error {
	var brd board

	var buf bytes.Buffer

	dump := func() {
		buf.Reset()
		brd.dumpInto(&buf)
		buf.WriteByte('\n')
		buf.WriteTo(out)
	}

	defer func(t0 time.Time) {
		t1 := time.Now()
		log.Printf("took %v", t1.Sub(t0))
	}(time.Now())

	// part 1
	if *after != 0 {
		n := *after
		brd.scores = append(make([]uint8, 0, 2*(n+10)))
		brd.init(0, 1, 3, 7)

		if *verbose {
			dump()
		}
		for m := n + 10 + 1; len(brd.scores) < m; {
			brd.tick()
			if *verbose {
				dump()
			}
		}

		buf.Reset()
		for _, s := range brd.scores[n : n+10] {
			buf.WriteByte('0' + byte(s))
		}
		fmt.Printf("%s\n", buf.Bytes())
		return nil
	}

	// part 2
	if *search != "" {
		var pattern []uint8
		for i := 0; i < len(*search); i++ {
			c := (*search)[i]
			d := c - '0'
			if d > 9 {
				return fmt.Errorf("invalid -search %q", *search)
			}
			pattern = append(pattern, d)
		}

		brd.scores = make([]uint8, 0, 1024*1024*1024)
		brd.init(0, 1, 3, 7)

		i := brd.search(pattern)
		if *verbose {
			dump()
		}
		log.Printf("FOUND @%v", i)

		return nil
	}

	return errors.New("what do you want? pass -after for part1 or -search for part2")
}

func (brd *board) init(e1, e2 int, scores ...uint8) {
	brd.e1, brd.e2 = e1, e2
	brd.scores = append(brd.scores, scores...)
}

func (brd *board) tick() {
	s1, s2 := brd.scores[brd.e1], brd.scores[brd.e2]
	tot := s1 + s2
	d1, d2 := tot/10, tot%10
	if tot < 10 {
		brd.scores = append(brd.scores, d2)
	} else {
		brd.scores = append(brd.scores, d1, d2)
	}
	brd.e1 = (brd.e1 + 1 + int(s1)) % len(brd.scores)
	brd.e2 = (brd.e2 + 1 + int(s2)) % len(brd.scores)
}

func (brd *board) search(pattern []uint8) int {
	sc := make([]uint8, len(pattern))
	ii := 0
	for i := 0; ; i++ {
		brd.tick()
		for ; ii < len(brd.scores)-len(sc); ii++ {
			copy(sc, brd.scores[i:])
			ok := true
			for j, v := range pattern {
				if sc[j] != v {
					ok = false
					break
				}
			}
			if ok {
				return i
			}
		}
	}
}

func (brd *board) dumpInto(buf *bytes.Buffer) {
	buf.Grow(len(brd.scores)*3 + 1)
	for i, s := range brd.scores {
		if i == brd.e1 && i == brd.e2 {
			fmt.Fprintf(buf, "<%d>", s)
		} else if i == brd.e1 {
			fmt.Fprintf(buf, "(%d)", s)
		} else if i == brd.e2 {
			fmt.Fprintf(buf, "[%d]", s)
		} else {
			fmt.Fprintf(buf, " %d ", s)
		}
	}
}

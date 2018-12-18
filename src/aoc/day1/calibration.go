package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

var deltas []int

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	if err := slurp(os.Stdin); err != nil {
		return err
	}

	counts := make(map[int]int, 2*len(deltas))

	rounds := 1

	total := 0
	for rounds < 10000 {
		for _, delta := range deltas {
			total += delta
			c := counts[total]
			c++
			counts[total] = c
			if c > 1 {
				log.Printf("FIRST: %v after %v rounds", total, rounds)
				return nil
			}
		}
		rounds++
	}

	return fmt.Errorf("found none in %v rounds", rounds)
}

func slurp(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		n, err := strconv.Atoi(sc.Text())
		if err != nil {
			return fmt.Errorf("failed to parse %q: %v", sc.Text(), err)
		}
		deltas = append(deltas, n)
	}
	return sc.Err()
}

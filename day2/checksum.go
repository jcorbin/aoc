package main

import (
	"bufio"
	"io"
	"log"
	"os"
)

func main() {
	if err := find(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

func checksum(r io.Reader) error {
	twos, threes := 0, 0
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var counts [256]uint8

		token := sc.Text()
		for i := 0; i < len(token); i++ {
			counts[token[i]]++
		}

		// log.Printf("token %q", token)
		// for i, c := range counts {
		// 	if c > 1 {
		// 		log.Printf("count %q => %v", byte(i), c)
		// 	}
		// }

		var have2, have3 int
		for i := 0; i < len(counts) && (have2 == 0 || have3 == 0); i++ {
			switch counts[i] {
			case 2:
				have2 = 1
			case 3:
				have3 = 1
			}
		}
		// log.Printf("have %v %v", have2, have3)
		twos += have2
		threes += have3

	}
	if err := sc.Err(); err != nil {
		return err
	}

	log.Printf("twos: %v threes: %v", twos, threes)
	log.Printf("checusum %v", twos*threes)

	return nil
}

func find(r io.Reader) error {
	table := make(map[string][]string, 64*1024)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		token := sc.Text()
		for i := 0; i < len(token); i++ {
			key := token[:i] + token[i+1:]
			table[key] = append(table[key], token)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	log.Printf("%v entries in table", len(table))

	type seenKey struct {
		a, b string
	}
	seen := make(map[seenKey]struct{}, 2*len(table))

	for tk, ent := range table {
		if len(ent) < 2 {
			continue
		}

		for i := 0; i < len(ent); i++ {
			for j := 0; j < len(ent); j++ {
				if i == j {
					continue
				}
				sk := seenKey{ent[i], ent[j]}
				if sk.b == sk.a {
					continue
				}
				if sk.b < sk.a {
					sk.a, sk.b = sk.b, sk.a
				}
				if _, checked := seen[sk]; checked {
					continue
				}
				seen[sk] = struct{}{}
				ndiff, ok := countDiff(sk.a, sk.b)
				if !ok {
					continue
				}
				if ndiff == 1 {
					log.Printf("a %v", sk.a)
					log.Printf("b %v", sk.b)
					log.Printf("k %v", tk)
				}
			}
		}
	}

	return nil
}

func countDiff(a, b string) (ndiff int, ok bool) {
	if len(a) != len(b) {
		return 0, false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			ndiff++
		}
	}
	return ndiff, true
}

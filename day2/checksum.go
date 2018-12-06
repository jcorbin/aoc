package main

import (
	"bufio"
	"io"
	"log"
	"os"
)

func main() {
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

func run(r io.Reader) error {
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

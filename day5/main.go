package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

func run(r io.Reader) error {
	// read chain
	chain, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// truncate chain
	for i, b := range chain {
		if b < 0x40 || 0x7f <= b {
			chain = chain[:i]
			break
		}
	}

reduce:
	for {

		log.Printf("%v units", len(chain))
		for i, j := 0, 1; j < len(chain); i, j = j, j+1 {
			if chain[i] != chain[j] && chain[i]^0x20 == chain[j] {
				log.Printf("Reduce @%v %q <-> %q", i, chain[i], chain[j])
				n := copy(chain[i:], chain[j+1:])
				chain = chain[:i+n]
				continue reduce
			}
		}

		break
	}

	fmt.Printf("final chain:\n%s\n", chain)

	return nil
}

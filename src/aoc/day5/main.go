package main

import (
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

	tmp := append([]byte(nil), chain...)
	tmp = react(tmp, nil /* log.Printf */)
	log.Printf("initial chain reduced to %v", len(tmp))

	var (
		best    byte
		bestLen int
	)
	for b := byte(0x40); b < 0x60; b++ {
		tmp = tmp[:len(chain)]
		copy(tmp, chain)
		tmp = prune(tmp, b)
		tmp = react(tmp, nil)
		if best == 0 || bestLen > len(tmp) {
			best, bestLen = b, len(tmp)
		}
	}

	log.Printf("reduced to %v by pruning %q", bestLen, best)

	return nil
}

func prune(chain []byte, b byte) []byte {
	B := b ^ 0x20
	i := 0
	for j := 0; j < len(chain); j++ {
		if chain[j] != b && chain[j] != B {
			chain[i] = chain[j]
			i++
		}
	}
	return chain[:i]
}

func react(chain []byte, logf func(string, ...interface{})) []byte {
reduce:
	for {
		if logf != nil {
			logf("%v units", len(chain))
		}
		for i, j := 0, 1; j < len(chain); i, j = j, j+1 {
			if chain[i] != chain[j] && chain[i]^0x20 == chain[j] {
				if logf != nil {
					logf("Reduce @%v %q <-> %q", i, chain[i], chain[j])
				}
				n := copy(chain[i:], chain[j+1:])
				chain = chain[:i+n]
				continue reduce
			}
		}
		return chain
	}
}

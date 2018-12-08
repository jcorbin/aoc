package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin))
}

type builder struct {
	ns []int
	i  int
}

func (bld *builder) next() (int, error) {
	if bld.i >= len(bld.ns) {
		return 0, errors.New("no more numbers")
	}
	r := bld.ns[bld.i]
	bld.i++
	return r, nil
}

func (bld *builder) build() (n node, _ error) {
	nc, err := bld.next()
	if err != nil {
		return n, err
	}
	nm, err := bld.next()
	if err != nil {
		return n, err
	}

	n.children = make([]node, 0, nc)
	n.metadata = make([]int, 0, nm)
	for len(n.children) < cap(n.children) {
		c, err := bld.build()
		if err != nil {
			return n, err
		}
		n.children = append(n.children, c)
	}
	for len(n.metadata) < cap(n.metadata) {
		v, err := bld.next()
		if err != nil {
			return n, err
		}
		n.metadata = append(n.metadata, v)
	}
	return n, nil
}

func run(r io.Reader) error {
	ns, err := readNums(r)
	if err != nil {
		return err
	}

	bld := builder{ns: ns}
	tree, err := bld.build()
	if err != nil {
		return fmt.Errorf("failed to build tree: %v", err)
	}

	// log.Printf("ns: %v", bld.ns)
	// log.Printf("tree: %v", tree)

	total := 0
	q := []node{tree}
	for len(q) > 0 {
		n := q[len(q)-1]
		q = q[:len(q)-1]
		q = append(q, n.children...)
		for _, m := range n.metadata {
			total += m
		}
	}
	log.Printf("total metadata: %v", total)

	return nil
}

type node struct {
	children []node
	metadata []int
}

func readNums(r io.Reader) (ns []int, _ error) {
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanWords)
	for sc.Scan() {
		token := sc.Text()
		n, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid token %q: %v", token, err)
		}
		ns = append(ns, n)
	}
	return ns, sc.Err()
}

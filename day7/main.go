package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/jcorbin/anansi"
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin))
}

func run(r io.Reader) error {
	reqs, err := readReqs(r)
	if err != nil {
		return err
	}

	type set map[string]struct{}
	type graph map[string]set

	// load dependency graph
	nodes := make(set, len(reqs))
	deps := make(graph, len(reqs))
	codeps := make(graph, len(reqs))
	for _, r := range reqs {
		nodes[r.a] = struct{}{}
		nodes[r.b] = struct{}{}
		if deps[r.a] == nil {
			deps[r.a] = set{r.b: struct{}{}}
		} else {
			deps[r.a][r.b] = struct{}{}
		}
		if codeps[r.b] == nil {
			codeps[r.b] = set{r.a: struct{}{}}
		} else {
			codeps[r.b][r.a] = struct{}{}
		}
	}

	// prune nodes
	for id := range nodes {
		if len(deps[id]) > 0 {
			delete(nodes, id)
		}
	}

	// topo sort
	var buf bytes.Buffer
	for len(nodes) > 0 {
		var id string
		for nid := range nodes {
			if id == "" || nid < id {
				id = nid
			}
		}
		buf.WriteString(id)
		delete(nodes, id)

		for nid := range codeps[id] {
			need := deps[nid]
			delete(need, id)
			if len(need) == 0 {
				delete(deps, nid)
				nodes[nid] = struct{}{}
			}
		}
		delete(codeps, id)
	}
	buf.WriteRune('\n')
	_, err = buf.WriteTo(os.Stdout)

	return err
}

type req struct {
	a, b string
}

var reqPattern = regexp.MustCompile(`^Step (\w+) must be finished before step (\w+) can begin\.$`)

func readReqs(r io.Reader) (reqs []req, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := reqPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			log.Printf("NO MATCH %q", line)
			continue
		}
		reqs = append(reqs, req{parts[2], parts[1]})
	}
	return reqs, sc.Err()
}

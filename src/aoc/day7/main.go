package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jcorbin/anansi"
)

var (
	numWorkers = flag.Int("n", 0, "number of workers")
	workBase   = flag.Int("w", 0, "base amount of work per task")
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

	deps := makedepgraph(len(reqs))
	deps.load(reqs)
	deps.pruneNodes()

	// part 1, print order
	if *numWorkers == 0 {
		var buf bytes.Buffer
		for len(deps.nodes) > 0 {
			id := deps.minNode()
			buf.WriteString(id)
			deps.finish(id)
		}
		buf.WriteRune('\n')
		_, err = buf.WriteTo(os.Stdout)
		return err
	}

	// part 2, schedule
	type work struct {
		id  string
		end int
	}
	workers := make([]work, *numWorkers)
	var done []string

	var buf bytes.Buffer
	buf.Grow(1024)

	const sep = ","

	fmt.Fprintf(&buf, "Second%s", sep)
	for i := range workers {
		fmt.Fprintf(&buf, "Worker %d%s", i+1, sep)
	}
	fmt.Fprintf(&buf, "Done\n")
	if _, err = buf.WriteTo(os.Stdout); err != nil {
		return err
	}

	out := func(t int) error {
		buf.Reset()
		fmt.Fprintf(&buf, "%d%s", t, sep)
		for _, wk := range workers {
			id := wk.id
			if id == "" {
				id = "."
			}
			fmt.Fprintf(&buf, "%s%s", id, sep)
		}
		fmt.Fprintf(&buf, "%s\n", strings.Join(done, ""))
		_, err = buf.WriteTo(os.Stdout)
		return err
	}

	wip := make(set)

	for t := 0; len(wip)+len(deps.co) > 0; t++ {
		for i, wk := range workers {
			if wk.id != "" && wk.end <= t {
				done = append(done, wk.id)
				deps.finish(wk.id)
				delete(wip, wk.id)
				workers[i] = work{}
			}
		}

		for i, wk := range workers {
			if wk.id == "" {
				if id := deps.minNode(); id != "" {
					delete(deps.nodes, id)
					wip[id] = struct{}{}
					workers[i] = work{
						id:  id,
						end: t + 1 + cost(id),
					}
				}
			}
		}

		if err := out(t); err != nil {
			return err
		}
	}

	return nil
}

func cost(id string) int {
	if len(id) != 1 {
		panic("bad string")
	}
	v := int(id[0] - 'A')
	return *workBase + v
}

type set map[string]struct{}
type graph map[string]set

type depgraph struct {
	nodes set
	dep   graph
	co    graph
}

func makedepgraph(n int) depgraph {
	return depgraph{
		nodes: make(set, n),
		dep:   make(graph, n),
		co:    make(graph, n),
	}
}

func (deps depgraph) load(reqs []req) {
	for _, r := range reqs {
		deps.nodes[r.a] = struct{}{}
		deps.nodes[r.b] = struct{}{}
		if deps.dep[r.a] == nil {
			deps.dep[r.a] = set{r.b: struct{}{}}
		} else {
			deps.dep[r.a][r.b] = struct{}{}
		}
		if deps.co[r.b] == nil {
			deps.co[r.b] = set{r.a: struct{}{}}
		} else {
			deps.co[r.b][r.a] = struct{}{}
		}
	}
}

func (deps depgraph) pruneNodes() {
	for id := range deps.nodes {
		if len(deps.dep[id]) > 0 {
			delete(deps.nodes, id)
		}
	}
}

func (deps depgraph) minNode() (id string) {
	for nid := range deps.nodes {
		if id == "" || nid < id {
			id = nid
		}
	}
	return id
}

func (deps depgraph) finish(id string) {
	delete(deps.nodes, id)
	for nid := range deps.co[id] {
		need := deps.dep[nid]
		delete(need, id)
		if len(need) == 0 {
			delete(deps.dep, nid)
			deps.nodes[nid] = struct{}{}
		}
	}
	delete(deps.co, id)
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

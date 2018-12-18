# [Advent of Code](https://adventofcode.com/2018)

## Setup

This repository is a [Go](https://golang.org/) monorepo, to build/test within
it, set `GOPATH` to the path to this repository. Use
[direnv](https://direnv.net) to automate this.

## Tests

Run `make test` to run all tests, and linters.

To test an individual day, run `go test aoc/dayN`.

## Building

To build all days, run `make days`.

To build a single day, run `make dayN` (or `go build aoc/dayN` if you prefer).

## Running

While each day differs slightly in detail, in general each `dayN` binary will:

- read problem input `os.Stdin`
- log results to `os.Stderr` and/or display `os.Stdout`
- some days run in fullscreen interactive mode (e.g. `day13` and `day15`)
- a few days take input parameters simply as command line flags

## Input

Each day's description is in `src/aoc/dayN/README.md`.

Most days have an example input in `.../dayN/ex.in` and my full problem input
in `.../dayN/input`.

Some days have their examples turned into test cases in `.../dayN/*_test.go`;
some also have benchmark cases.

# [Advent of Code](https://adventofcode.com/2018)

## Days

| Day                              | Description                                   |
|----------------------------------|-----------------------------------------------|
| [day1](src/aoc/day1/README.md)   | Sum numbers (basically a mic test)            |
| [day2](src/aoc/day2/README.md)   | Token checksumming, by counting letters       |
| [day3](src/aoc/day3/README.md)   | Counting rectangle overlap                    |
| [day4](src/aoc/day4/README.md)   | Guard log parsing and schedule analysis       |
| [day5](src/aoc/day5/README.md)   | String reduction by adjacent `aA` collapse    |
| [day6](src/aoc/day6/README.md)   | Voronoi diagram under Manhattan distance      |
| [day7](src/aoc/day7/README.md)   | Dependency ordering, and scheduling           |
| [day8](src/aoc/day8/README.md)   | Tree decoding and walking                     |
| [day9](src/aoc/day9/README.md)   | Marble Ring Game                              |
| [day10](src/aoc/day10/README.md) | Particle simulation to spell letters          |
| [day11](src/aoc/day11/README.md) | Generative 2D grid point and region finding   |
| [day12](src/aoc/day12/README.md) | 1D Cellular Automata                          |
| [day13](src/aoc/day13/README.md) | Minecart game                                 |
| [day14](src/aoc/day14/README.md) | Chocolate recipe (number sequence expansion)  |
| [day15](src/aoc/day15/README.md) | Goblin v Elf combat game sim                  |
| [day16](src/aoc/day16/README.md) | Machine Code Reversing                        |
| [day17](src/aoc/day17/README.md) | Wateflow sim                                  |
| [day18](src/aoc/day18/README.md) | 2D Cellular Automata                          |
| [day19](src/aoc/day19/README.md) | Machine code run and analysis                 |
| [day20](src/aoc/day20/README.md) | Regex room & door map gen                     |
| [day21](src/aoc/day21/README.md) | Machine code analysis                         |
| [day22](src/aoc/day22/README.md) | Cave rescue maze                              |
| [day23](src/aoc/day23/README.md) | Nanobot 3D space + radius analysis            |
| [day24](src/aoc/day24/README.md) | Immune system combat sim                      |
| [day25](src/aoc/day25/README.md) | 4D space analysis                             |

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

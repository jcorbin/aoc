# [Advent of Code 2024](https://adventofcode.com/2024)

| 📆                 | 🏷️                                                   |📜|
|--------------------|------------------------------------------------------|--|
| [1][d1]            | Ranked variations and similarity freqing             |🌟|
| [2][d2]            | Damp level deltas                                    |🌟|
| [3][d3]            | Simple mechanistic mulling                           |🌟|
| [4][d4]            | Word Searching                                       |⭐|
| [2015-19][d201519] | Transitional molecules                               |⭐|
| [2020-02][d202002] | Password policy validation                           |🌟|
| 2019-10            | TODO back ref from day 4                             |⌛|

## Running

To run a given day:
```shell
$ python run.py <day> <part> [<input>]
```
- `day` is number
- `part` is number, probably just `1` or `2`
- `input` is file, defaults to `dayD/partP/input.txt` ; may give `-` for stdin

There are also likely test cases for each day:
```shell
$ pytest day1/part1.py
```

# Dev Log
## 2024-12-02

Guess we're doing this thing, let's see how far we get, probably going to go with low effort python code tbh.

### [Day 1][d1]

Started out by cribbing the `strkit.py` module from recent fascination with
[tool-assisted word game play](https://github.com/jcorbin/alphahack).

The primary thing this gives is a rather ergonomic way to shovel AOC's
conversational test examples into actual test cases.

Structurally am going with a similar layout to my 2022 zig run:
- folder for each day
- self contained part1/part2 implementations
- no code sharing between days, just copy modules into each day

From prior years, that point about avoiding code sharing can be important, as
the inevitable urge to re-use code from prior day will run smack into one's
impulse to evolve said code, breaking prior day solutions.

Since this is ~~Just A Wendy's~~ python, immediately hit a snag around
relative import resolution:
- pyright really wanted to see `import .strkit`
- which conflicts with also having an inline `__name__ == 'main'`
- instead hoisting the main harness up out of each day/part let me dodge further
  package system boilerplate/reorganization, plus ended up with better
  ergonomics than a more raw main stub inline each day

### [Day 2][d2]

- evolved toplevel `run.py` to pass a `-v` arg wired `verbose` flag along to
  each problem/solution
- ...may eventually prove useful for prompt file introspection, e.g. scrape test
  fixture data, or for doing some kinda more interactive harness
- overall another simple starter day, so happy to continue settle into the
  `MarkedSpec` test and `run.py` harness groove
- ...there's something altogether soothing about:
  1. transliterating upstream problem statement into markdown
  2. writing test specs in markdown-ish
  3. generating optionally verbose output in a similar streamed-text form

### [2015 Day 19][d201519]

- backtracking a reference from [day 2][d2]
- further evolved `run.py` evolution to support out-of-sequence / prior-year days
- managed to solve the part 1 calibration easily enough, but trying to just low
  effort brute force the search in part 2 didn't work out
- curious if this will become topical this year
- may try a reverse search re-attempt to part 2, or just trying to make my brute
  searcher faster, by say not using regex to deconstruct each molecule string

## 2024-12-03

### [Day 3][d3]

- cute pattern matching problem in part 1, managed to smoke out an edge case bug in `strkit.spliterate`
- had another back reference to 2020-02, so may bet to that after part 2
- curious to see if part 2 expands on valid commands or what...
- ...yep nothing that a couple more legs in my instruction matching regex can't handle with ease

### [2020 Day 2][d202002]

- back reference from [day 3][d3]
- easy part1, just parse, count, check, done.
- easy part2, done done.
- no notes really, just easy string mucking.

## 2024-12-03

### [Day 4][d4]

- part 1: so while this one seemed simple at first read, decided to go with a
  sort of djsktra-map style approach, rather than something like a recursive
  letter checker
  - ... which took way longer than expected to get correct, let's hope that pays
    off in part 2...
  - on the upside, solution was very fast, even on larger input grid
- part 2: nope, djsktra-map approach was irrelevant
  - so wiped out most of prior code, replaced with a simple minded stencil...
  - ...passes test case, but fails the part 2 prompt
  - may try more later, but have exhausted my "low effort" budget for now

[d1]: day1/
[d2]: day2/
[d3]: day3/
[d4]: day4/
[d201519]: back_2015_day19/
[d202002]: back_2020_day2/

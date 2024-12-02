# [Advent of Code 2024](https://adventofcode.com/2024)

| 📆          | 🏷️                                                          |📜|
|-------------|-------------------------------------------------------------|--|
| [1](day1/)  | Ranked variations and similarity freqing                    |🌟|
| [2](day2/)  | Damp level deltas                                           |🌟|

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

### [Day 1](day1/)

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

### [Day 2](day2/)

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

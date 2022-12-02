# 2022-12-02

## Day 2

- TIL that zig uses `and` / `or` instead of `&&` / `||`
- hanging methods off enums (for move and outcome modeling) feels good, even
  better than it did in go
- curious to investigate more how the unfamiliar integer bit widths work (like
  the `u2` used to back our enums today)
- TODO: if/capture staircases don't vibe

# 2022-12-01

- got zig hello world program working:
  - this will be my first time trying to write zig, after having skimmed their
    docs a bit over the last few weeks
  - editor setup was easy: neovim already had the treesitter definition, a
    filetype plugin, and an lsp config definition for `zls`
  - not going to be at all competitive or especially punctual about solutions
    this year, since my main goal is just to learn zig
- merged with my prior 2018 `github.com/jcorbin/aoc` repository

## Day 1

Went well for learning zig from scratch, zig thoughts:

- already feels like a familiar "better C/Go", immediately addressing many of
  my top 😫s from Go:
  - much better error handling
  - type system seems to greatly simplify things like reader orchestration so
    far; since the std reader type can be concrete, it can carry all the
    batteries, rather than needing other ancillary types/structs to implement
    things like delimiter scan.
  - slices are much less magical
  - strings are way better, and avoids Go's well known byte-slice vs string hell
  - array repetition by exponentiation is a nice take
  - the if/while binding deal will take a bit of getting used to, but actually
    already feels better than Go's `if x := ...; x` affair, if maybe less flexible?
  - TODO: would like to revisit this solution, and unify the part1/part2
    implementations around some group-sum scanning utility
  - TODO: using the heap page allocator seems unnecessary here, would prefer a
    stack or arena allocator

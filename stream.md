# 2022-12-05: Day 5

WIP

# 2022-12-04: Day 4

- light day, no surprises, just the subtlety of implementing range overlap
- still ruminating on how to abstract my approach to advent problem harnessing
- no big zig learnings today, just minor ones:
  - getting used to implementing structs with methods using `@This()`
  - spent a bit of time figuring out how struct literals work, kept trying to
    write `.{ a, b }` positionally rather than nominal `.{ .A = a, .B = b }`

# 2022-12-03: Day 3

- took the oppurtunity to learn zig testing today; 10/10 would recommend
  - tests alongside their actual code as tighly as possible feels good
- but using `anytype` to build a function that can take any reader / writer
  doesn't; zig seesm to lack any way to abstract an interface
  - the only real drawback of this currently seems to be developer
    experience, since tooling can't know anything beyond "🤷 it's anytype"
  - so no code completion for "wait, what's that long-af method called to scan
    a line? Right, right, `readUntilDelimiterOrEof`, how silly of me, there we
    go..."
- for this problem, a better solution may have been to abstract around an input
  "line iterator" function, and an output "line consumer" function
- while I had fun golfing down the bit representations of each rucksack piece
  by using a code-indexed `u2` (part 1) and `u3` (part 2)...
  - ... I'm now wondering if zig's vector types would've done better; not sure,
    haven't tried to use that aspect of the language yet
- I tried again, and failed again, to write an iterable function (function that
  returns an iterator, iterator being a struct with a next method) for part 1
- defined my first error set, which feels good; however I'd like to see if
  there are any better way to provide context with an error (what was the
  problematic data? any prior conflicting data?); currently the only thing I've
  done is just print such data...

# 2022-12-02: Day 2

- TIL that zig uses `and` / `or` instead of `&&` / `||`
- hanging methods off enums (for move and outcome modeling) feels good, even
  better than it did in go
- curious to investigate more how the unfamiliar integer bit widths work (like
  the `u2` used to back our enums today)
- learned about the `orelse` way to unwrap nulls

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

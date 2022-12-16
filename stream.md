# 2022-12-16

# Day 15

- last night's issue wasn't actually a bug, just a wart around not merging a
  nil-gap away when `a.end == b.start`
- otoh took a nice diversion into lexically labeled area display, and spent too
  much time trying to make an x-axis ruler
- TODO today's final part 2 solution exhibits WRONG behavior under any
  non-Debug build ; circle back and understand how I'm abusing UB, or chase a
  report upstream otherwise

# 2022-12-15

# Day 15

- got most of the way done, but ran out of gas trying to debug range gap issues

# Day 14

- exported copy of grid module to new `/lib`
- updated template main a little

# 2022-12-14

# Day 14

- very enjoyable, and brings back 2018 simulation memories for sure
- almost went to a sparse or sharded world for part 2, but resisted the urge
  and just computed the likely source -> floor cone
- only taking around 400-700ns to pour each grain of sand feels good!

# Day 13

- learning more about data structure design trade offs: kept trying to use a
  singly or doubly linked list, switched to array list in the end for sake of
  simplicity
- comparison implemented over a flattening iterator was a neat take
- updated template
  - simplified parser: dropped `expect*` methods since narrowing their returned
    error type is either not obvious or not possible
  - variable verbosity level
- TODO try binary search for the find-key phase in part 2
- TODO why is parsing so slow? is it the array lists or just the depth of graph?

# 2022-12-13

- template progress
  - adopted builder init/parse/finish -> world from day12
  - main general page allocator
  - arg parsing arena
  - reified timer collection

## Day 12

- finally finished, but running out of focus, probably goint to defer Day 13 to
  tomorrow
- crash was due to using raw heap page allocator with multi array list... it
  tried to free a sub-region
  - got to learn about systemd coredump collection tho, which is super nice

# 2022-12-12

- updates for breaking changes in zig v0.10
  - template and Day 11 updated
  - TODO marshal all prior day tests, and update

## Day 12

- hit an interesting uncreachable crash with OOM preventing stack dump
- ran out of steam to finish today, will debug tomorrow

# 2022-12-11: Day 11

- whew... I kept trying to use big integers too long, before breaking down and
  getting smart about it...

No self reference by value:
```zig
/// A simple math expression on a single variable X
const Op = union(enum) {
    x: void,
    value: u32,
    add: [2]@This(),
    mul: [2]@This(),
};
```

Yes self reference by pointer:
```zig
/// A simple math expression on a single variable X
const Op = union(enum) {
    x: void,
    value: u32,
    add: [2]*@This(),
    mul: [2]*@This(),
};
```

Yes self reference by slice, which is a pointer+len:
```zig
/// A simple math expression on a single variable X
const Op = union(enum) {
    x: void,
    value: u32,
    add: []@This(),
    mul: []@This(),
};
```

## Day 10

- export `parse.Cursor` int parsing improvement back to template

# 2022-12-10

## Day 10

- arg parser paid off today, as I ended up adding double verbosity, and several
  tiers of modal functionality
- the before / during / after cycle semantics were the part that tripped me up
  most
- not happy with how janky the `CPU` / `CRT` struct integration is, and also wrt
  `Signal` for part 1
  - ... but I'm very happy with around 40ns typical per-line evaluation time
  - ... and total runtime around 30us

## Day 9

- exported progress to template, updating perf module
  - table oriented test block, eases multiple rounds of testing (e.g. if part 2
    is modally different from part 1, rather than an extension)
  - added app `Config` struct, part of test case definition, and now filled in
    by arg parsing in `main()`
    - template provides post per-line parse verbosity support using this
- further developed arg parsing module thru template

# 2022-12-09

## Day 9

- finally busted out the big guns today and used a HashMap for the first time
  to index spatial data
  - TODO: now I'm really itching to write a linear quadtree in zig... sooon
- finally broke down and threaded an application config struct thru, allowing
  part 1 vs 2 differences to be parameterized today
  - used by test case enumeration
  - used by hacky process argument parsing
- really appreciated step-by-step testing today
- furthest I've gone towards implementing a game style simulation world this
  year

## Day 8

- exported progress to template, updating perf module

# 2022-12-08

## Day 8

- evolved day 7 perf module
- wrote part 1 straight ahead, then factored out a Point utility for part 2
- TODO: would like to learn how to parse argv/flags, since printing the
  visibility field is almost the most expensive phase
- TODO: would like to evaluate alternate array layout like z-order
- TODO: maybe try an alternate algorithm for part 2
  - consider marching a `[9]usize` tier-count array thru the same pattern as
    part 1 (from each edge)
  - multiply into a `score` dimension

## Day 7

- exported progress to template, including perf module

# 2022-12-07: Day 7

- today took longer than expected
  - invested in evolving the parser module from day 5
    - spent too long chasing my own tail trying to be too general, but not
      having enough zig experience to really pull that off yet
  - this was the first time I've really had to copy/retain input data in
    meaningful ways
    - so I spent too long futzing around with my own Pool notions
    - before just using an arena allocator
    - but I hit some early segfaults when doing so:
      - turns out that `Allocator.init()`
      - which is called via `ArenaAllocator.allocator()` must be called on a
        stable copy of the arena
      - not too far down the stack, since arenas are stored / passed by value usually
      - say in `MyType.init(allocator)` trying to create a retained arena
- however I did succeed in doing some tree-walk iteration, so my comfort level
  with zig seems to be coming along nicely
- TODO would like to evaluate optimizations likes:
  - slab/pool allocating things, even under the arenas we have now
  - this might be even more of a win for `DirWalker`, since we do 3 traversals
    in the end, with no node reuse between them
  - generally learn how to profile zig code, and see where's the beef
- wrote some perf timing code afterwards, looks works in <1uSec with most time
  spent parsing/buliding the device filesystem

# 2022-12-06

## Day 6

- all that setup, and then today needed neither allocation nor line based
  reading; this only stands to reason
- directly coded comparisons in part 1, hoped that I could use a inlined loop
  for part 2, but wasn't able to
- also found out that sub tests don't seem to be possible

## Day 5

- factored out parse and slab modules within the day for potential copy into
  future days (want to avoid direct inter-day code sharing to avoid needing to
  update any prior days as things evolve)
- updated template
  - debug print run output on test failure
  - provide an allocator to run, and setup an arena within
  - provide a stub line reading loop

# 2022-12-05: Day 5

- dug in on allocation and parsing games today
  - for whatever reason, I wasn't able to use an arena without leaking in the
    test, so I implemented a simpler slab (of list nodes) deal
  - the unified line parser feels great actually; needs more work before being
    really ergonomic, but works well enough to already be more coherent than a
    network of `split`/`tokenize` and `parseInt`
- started building an inter-day template based on day 4
  - TODO update template based on day 5 progress
- TODO factor out the `ParseCursor` deal into a module
- TODO factor out the `SlabChain` deal into a module

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

#!/usr/bin/env python

import argparse
import importlib
import re
import sys
from collections.abc import Generator, Iterable
from typing import cast, final, Callable

@final
class Solution:
    def __init__(self, entry: Callable[..., object], verbose: int = 0):
        self.entry = entry
        self.verbose = verbose

    def run(self, input: str) -> Generator[object]:
        if input.lower() in ('-', '/dev/stdin', '<stdin>'):
            yield from self.wrap_output(sys.stdin)
        else:
            with open(input) as f:
                yield from self.wrap_output(f)

    def wrap_output(self, input: Iterable[str]) -> Generator[object]:
        res = self.entry(input, verbose=self.verbose)
        if isinstance(res, Iterable):
            yield from res
        else:
            yield res

@final
class Problem:
    def __init__(self, day: str, part: int = 1):
        self.day = day
        self.part = part
        self.module = importlib.import_module(f'{day}.part{part}')

    @property
    def entry(self):
        entry = cast(object, getattr(self.module, 'run'))
        if not callable(entry):
            raise ValueError(f'no run() entry point found in {self.module!r}')
        return entry

    @property
    def day_input(self):
        return f'{self.day}/input.txt'

    def solve(self, *, inputname: str|None = None, verbose: int = 0):
        pr = Solution(self.entry, verbose=verbose)
        return pr.run(inputname or self.day_input)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    _ = parser.add_argument('day')
    _ = parser.add_argument('part', nargs='?', default=1, type=int)
    _ = parser.add_argument('input', nargs='?', default='')
    _ = parser.add_argument('-v', default=0, action='count')
    args = parser.parse_args()

    day = cast(str, args.day)
    if re.match(r'\d+$', day):
        day = f'day{day}'

    pr = Problem(day, cast(int, args.part))
    for line in pr.solve(
        inputname = cast(str, args.input),
        verbose = cast(int, args.v),
    ): print(line)

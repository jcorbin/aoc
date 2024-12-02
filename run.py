#!/usr/bin/env python

import argparse
import importlib
import sys
from collections.abc import Generator, Iterable
from typing import cast, Callable

def load(day: str, part: str):
    impl = importlib.import_module(f'{day}.{part}')

    entry = cast(object, getattr(impl, 'run'))
    if not callable(entry):
        print('! no run() entry point found in {impl!r}')
        sys.exit(1)

    return entry

def run_input(entry: Callable[..., object], input: str) -> Generator[object]:
    if input.lower() in ('-', '/dev/stdin', '<stdin>'):
        yield from run_output(entry, sys.stdin)
    else:
        with open(input) as f:
            yield from run_output(entry, f)

def run_output(entry: Callable[..., object], input: Iterable[str]) -> Generator[object]:
    res = entry(input)
    if isinstance(res, Iterable):
        yield from res
    else:
        yield res

def main():
    parser = argparse.ArgumentParser()
    _ = parser.add_argument('day', type=int)
    _ = parser.add_argument('part', type=int)
    _ = parser.add_argument('input', nargs='?')
    args = parser.parse_args()

    day = cast(int, args.day)
    part = cast(int, args.part)
    input = cast(str, args.input) or f'day{day}/input.txt'
    sol = load(f'day{day}', f'part{part}')

    res = run_input(sol, input)
    for line in res:
        print(line)

if __name__ == '__main__':
    main()

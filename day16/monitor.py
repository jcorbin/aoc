#!/usr/bin/env python

# pylint: disable=missing-module-docstring
# pylint: disable=missing-class-docstring
# pylint: disable=missing-function-docstring

from collections import deque
from datetime import timedelta
import os
import re
import sys

def sify(num):
    if num > 1e12:
        return f'{num/1e9}T'
    if num > 1e9:
        return f'{num/1e9}B'
    if num > 1e6:
        return f'{num/1e6}M'
    if num > 1e3:
        return f'{num/1e9}K'
    return f'{num}'

def analyze(lines):
    search_states = 0
    search_time = 0
    search_depth = 0
    for line in lines:
        if not line.strip():
            continue
        match = re.search(r'searched (\d+) in (\d+).* depth = (\d+)', line)
        if match:
            line_states = int(match.group(1))
            line_time = int(match.group(2))
            line_depth = int(match.group(3))

            search_states += line_states
            search_time += line_time
            search_depth = max(search_depth, line_depth)

        yield line.rstrip('\r\n')

    rate = round(search_states / (search_time / 1e9))
    avg = round(search_time / search_states)
    took = timedelta(microseconds=search_time/1e3)

    yield f'Total searched {sify(search_states)} in {took} max depth: {search_depth}'
    yield f'- rate:{rate} states/s avg:{avg}ns'

def main():
    try:
        lines = int(os.environ['LINES'])
    except KeyError:
        lines = os.get_terminal_size().lines

    for line in deque(analyze(sys.stdin), lines):
        print(line)

main()

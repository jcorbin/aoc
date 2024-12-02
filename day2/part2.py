import re
import pytest
from collections.abc import Generator, Iterable, Sequence
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #given
    > 7 6 4 2 1
    > 1 2 7 8 9
    > 9 7 6 2 1
    > 1 3 2 4 5
    > 8 6 4 4 1
    > 1 3 6 7 9
    // - `7 6 4 2 1`: **Safe** without removing any level.
    // - `1 2 7 8 9`: **Unsafe** regardless of which level is removed.
    // - `9 7 6 2 1`: **Unsafe** regardless of which level is removed.
    // - `1 3 2 4 5`: **Safe** by removing the second level, `3`.
    // - `8 6 4 4 1`: **Safe** by removing the third level, `4`.
    // - `1 3 6 7 9`: **Safe** without removing any level.
    //
    // Thanks to the Problem Dampener, **`4`** reports are actually **safe**!
    - verbose: yes
    - output: ```
    // Safe [7, 6, 4, 2, 1]
    // Unsafe [1, 2, 7, 8, 9]: all diffs must be at most 3
    // Unsafe [9, 7, 6, 2, 1]: all diffs must be at most 3
    // Unsafe [1, 3, 2, 4, 5]: neither all increasing nor all decreasing
    // Safe Dampened sans [1]=3
    // Unsafe [8, 6, 4, 4, 1]: neither all increasing nor all decreasing
    // Safe Dampened sans [2]=4
    // Safe [1, 3, 6, 7, 9]
    4
    ```

    #given_terse
    > 7 6 4 2 1
    > 1 2 7 8 9
    > 9 7 6 2 1
    > 1 3 2 4 5
    > 8 6 4 4 1
    > 1 3 6 7 9
    - output: 4

''')), ids=MarkedSpec.get_id)
def test(spec: MarkedSpec):
    # for line in spec.speclines:
    #     print('WAT', repr(line))

    lines = spliterate(spec.input, '\n')
    expected_output: list[str] = []
    verbose = False
    for name, value in spec.props:
        if name == 'output': expected_output.extend(spliterate(value, '\n'))
        elif name == 'verbose': verbose = any(value.lower().startswith(c) for c in 'ty')
        else: raise ValueError(f'invalid test prop {name!r}')
    assert list(run(lines, verbose=verbose)) == expected_output

def check(levels: Sequence[int]):
    # So, a report only counts as safe if both of the following are true:
    # - The levels are either **all increasing** or **all decreasing**.
    # - Any two adjacent levels differ by **at least one** and **at most three**.

    diffs = [
        b - a
        for b, a in zip(levels[1:], levels)]

    if not (
        all(diff < 0 for diff in diffs) or
        all(diff > 0 for diff in diffs)
    ): raise ValueError('neither all increasing nor all decreasing')

    diffs = [abs(diff) for diff in diffs]

    if any(diff < 1 for diff in diffs):
        raise ValueError('all diffs must be at least 1')

    if any(diff > 3 for diff in diffs):
        raise ValueError('all diffs must be at most 3')

def run(input: Iterable[str], verbose: bool=False) -> Generator[str]:
    pattern = re.compile(r'(?x) \d+ ( [ ]+ \d+ )* [ ]* $')

    res: int = 0
    for line in input:
        match = pattern.match(line)
        assert match

        levels = [
            int(token)
            for token in spliterate(line, ' ')]
        try:
            check(levels)
        except ValueError as e:
            if verbose:
                yield f'// Unsafe {levels!r}: {e}'
        else:
            if verbose:
                yield f'// Safe {levels!r}'
            res += 1
            continue

        for i in range(len(levels)):
            sans = [*levels[:i], *levels[i+1:]]
            try:
                check(sans)
            except ValueError as e:
                pass
            else:
                if verbose:
                    yield f'// Safe Dampened sans [{i}]={levels[i]}'
                res += 1
                break

    yield f'{res}'

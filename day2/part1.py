import re
import pytest
from collections.abc import Generator, Iterable
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #given
    > 7 6 4 2 1
    > 1 2 7 8 9
    > 9 7 6 2 1
    > 1 3 2 4 5
    > 8 6 4 4 1
    > 1 3 6 7 9
    // - `7 6 4 2 1`: **Safe** because the levels are all decreasing by `1` or `2`.
    // - `1 2 7 8 9`: **Unsafe** because `2 7` is an increase of `5`.
    // - `9 7 6 2 1`: **Unsafe** because `6 2` is a decrease of `4`.
    // - `1 3 2 4 5`: **Unsafe** because `1 3` is increasing but `3 2` is decreasing.
    // - `8 6 4 4 1`: **Unsafe** because `4 4` is neither an increase or a decrease.
    // - `1 3 6 7 9`: **Safe** because the levels are all increasing by `1`, `2`, or `3`.
    //
    // So, in this example, **`2`** reports are *safe*.
    - verbose: yes
    - output: ```
    // Safe [7, 6, 4, 2, 1]
    // Unsafe [1, 2, 7, 8, 9]: all diffs must be at most 3
    // Unsafe [9, 7, 6, 2, 1]: all diffs must be at most 3
    // Unsafe [1, 3, 2, 4, 5]: neither all increasing nor all decreasing
    // Unsafe [8, 6, 4, 4, 1]: neither all increasing nor all decreasing
    // Safe [1, 3, 6, 7, 9]
    2
    ```

    #given_terse
    > 7 6 4 2 1
    > 1 2 7 8 9
    > 9 7 6 2 1
    > 1 3 2 4 5
    > 8 6 4 4 1
    > 1 3 6 7 9
    - output: 2

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

def run(input: Iterable[str], verbose: bool=False) -> Generator[str]:
    pattern = re.compile(r'(?x) \d+ ( [ ]+ \d+ )* [ ]* $')

    res: int = 0
    for line in input:
        match = pattern.match(line)
        assert match

        # So, a report only counts as safe if both of the following are true:
        # - The levels are either **all increasing** or **all decreasing**.
        # - Any two adjacent levels differ by **at least one** and **at most three**.

        levels = [
            int(token)
            for token in spliterate(line, ' ')]

        diffs = [
            b - a
            for b, a in zip(levels[1:], levels)]

        if not (
            all(diff < 0 for diff in diffs) or
            all(diff > 0 for diff in diffs)
        ):
            if verbose:
                yield f'// Unsafe {levels!r}: neither all increasing nor all decreasing'
            continue

        diffs = [abs(diff) for diff in diffs]

        if any(diff < 1 for diff in diffs):
            if verbose:
                yield f'// Unsafe {levels!r}: all diffs must be at least 1'
            continue

        if any(diff > 3 for diff in diffs):
            if verbose:
                yield f'// Unsafe {levels!r}: all diffs must be at most 3'
            continue

        if verbose:
            yield f'// Safe {levels!r}'
        res += 1

    yield f'{res}'

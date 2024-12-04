import re
from collections.abc import Generator, Iterable
from typing import cast
from .strkit import spliterate, MarkedSpec

@MarkedSpec.mark('''

    #given
    > MMMSXXMASM
    > MSAMXMSMSA
    > AMXSXMAAMM
    > MSAMASMSMX
    > XMASAMXAMM
    > XXAMMXXAMA
    > SMSMSASXSS
    > SAXAMASAAA
    > MAMMMXMMMM
    > MXMXAXMASX
    - verbose: 0
    // In this example, an `X-MAS` appears **`9`** times.
    - output: 9

''')
def test(spec: MarkedSpec):
    expected_output: list[str] = []
    verbose = 0
    for name, value in spec.props:
        if name == 'output': expected_output.extend(spliterate(value, '\n', trim=True))
        elif name == 'verbose': verbose = int(value)
        else: raise ValueError(f'invalid test prop {name!r}')
    lines = spliterate(spec.input, '\n')
    have_output = list(run(lines, verbose=verbose))
    assert have_output == expected_output

def run(input: Iterable[str], verbose: int = 0) -> Generator[str]:
    pattern = re.compile(r'(?x) ( \w+ )')

    grid: list[str] = []
    width: int = 0
    height: int = 0
    for line in input:
        if not line.strip(): continue
        match = pattern.match(line)
        if not match:
            raise RuntimeError(f'invalid input {line!r}')

        row = cast(str, match.group(1))
        if not width:
            width = len(row)
        elif len(row) != width:
            raise RuntimeError(f'invalid row size {len(row)} expected {width}')

        height += 1
        grid.extend(row)

    count: int = 0

    word: str = 'MAS'

    stencil = tuple(
        y*width + x
        for y in range(len(word))
        for x in range(len(word)))

    erase = set(range(len(stencil)))
    for i in range(len(word)):
        j = i*len(word) + i
        k = i*len(word) + len(word) - 1 - i
        if j in erase: erase.remove(j)
        if k in erase: erase.remove(k)
    erase = tuple(erase)

    # TODO nice to generate
    valid = set([

        'M.M' +
        '.A.' +
        'S.S',

        'S.M' +
        '.A.' +
        'S.M',

        'M.S' +
        '.A.' +
        'M.S',

        'S.S' +
        '.A.' +
        'M.M'

    ])

    for y in range(height - len(word)):
        for x in range(height - len(word)):
            at = y*width + x
            ats = tuple(at + offset for offset in stencil)
            dat = list(grid[i] for i in ats)
            for i in erase: dat[i] = '.'
            s = ''.join(dat)

            if s in valid:
                if verbose:
                    yield f'// found {s}'
                count += 1

    yield f'{count}'

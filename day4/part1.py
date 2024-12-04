import re
from collections.abc import Generator, Iterable
from typing import cast
from .strkit import spliterate, MarkedSpec

@MarkedSpec.mark('''

    #given
    // The actual word search will be full of letters instead. For example:
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
    // In this word search, `XMAS` occurs a total of **`18`** times
    - output: 18

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
    word: str = 'XMAS'
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

    offsets = (
        (-1, -1),
        ( 0, -1),
        ( 1, -1),
        (-1,  0),
        ( 1,  0),
        (-1,  1),
        ( 0,  1),
        ( 1,  1),
    )

    # def show(dgrid: list[int]):
    #     for offset in range(0, len(grid), width):
    #         wrow = grid[offset:offset+width]
    #         drow = dgrid[offset:offset+width]
    #         yield f'//     {''.join(
    #             "." if not d else l
    #             for d, l in zip(drow, wrow)
    #         )}'

    count: int = 0
    for dx, dy in offsets:
        if verbose:
            yield f'// offset: {dx},{dy}'

        dgrid = [0 for _ in grid]

        for dlet, let in enumerate(word, 1):
            # yield f'// {let} {dlet}'

            prior = dlet-1

            if prior > 0:
                at = [
                    divmod(j, width)
                    for j, d in enumerate(dgrid)
                    if d == prior]
                for y, x in at:
                    y += dy
                    x += dx
                    if not (0 <= x < width): continue
                    if not (0 <= y < height): continue
                    dgrid[y*width + x] = prior

            for j, (d, c) in enumerate(zip(dgrid, grid)):
                if d == prior:
                    dgrid[j] = dlet if c == let else 0

            # yield from show(dgrid)

        if verbose:
            yield f'// found: {dgrid.count(4)}'
        # yield from show(dgrid)

        count += dgrid.count(4)

    yield f'{count}'

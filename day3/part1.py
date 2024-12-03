import re
import pytest
from collections.abc import Generator, Iterable
from typing import cast
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #given
    // For example, consider the following section of corrupted memory:
    > xmul(2,4)%&mul[3,7]!@^do_not_mul(5,5)+mul(32,64]then(mul(11,8)mul(8,5))
    - verbose: 1
    // Only the four highlighted sections are real `mul` instructions.
    - output: ```
    // found: mul(2,4)
    // found: mul(5,5)
    // found: mul(11,8)
    // found: mul(8,5)
    ```
    // Adding up the result of each instruction produces **`161`** ( `2*4 + 5*5 + 11*8 + 8*5` ).
    - output: 161

''')), ids=MarkedSpec.get_id)
def test(spec: MarkedSpec):
    expected_output: list[str] = []
    verbose = 0
    for name, value in spec.props:
        if name == 'output': expected_output.extend(spliterate(value, '\n', trim=True))
        elif name == 'verbose': verbose = int(value)
        else: raise ValueError(f'invalid test prop {name!r}')
    print('WAT', repr(spec.input))
    print('WOT', list(spliterate(spec.input, '\n')))
    lines = spliterate(spec.input, '\n')
    have_output = list(run(lines, verbose=verbose))
    assert have_output == expected_output

def run(input: Iterable[str], verbose: int = 0) -> Generator[str]:
    pattern = re.compile(r'(?x) mul \( ( \d{1,3} ) , ( \d{1,3} ) \)')

    res: int = 0
    for line in input:
        for match in pattern.finditer(line):
            if verbose:
                yield f'// found: {match.group(0)}'
            sa = cast(str, match.group(1))
            sb = cast(str, match.group(2))
            a = int(sa)
            b = int(sb)
            m = a * b
            res += m
            if verbose > 1:
                yield f'// res += ( {a} * {b} = {m} ) = {res}'

    yield f'{res}'

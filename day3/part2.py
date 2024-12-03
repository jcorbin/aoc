import re
import pytest
from collections.abc import Generator, Iterable
from typing import cast
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #given
    // For example:
    > xmul(2,4)&mul[3,7]!^don't()_mul(5,5)+mul(32,64](mul(11,8)undo()?mul(8,5))
    - verbose: 2
    // This corrupted memory is similar to the example from before,
    // but this time the `mul(5,5)` and `mul(11,8)` instructions are
    // **disabled** because there is a `don't()` instruction before them.
    // The other `mul` instructions function normally,
    // including the one at the end that gets re-**enabled** by a `do()` instruction.
    - output: ```
    // found: mul(2,4)
    // res += ( 2 * 4 = 8 ) = 8
    // found: don't()
    // found: mul(5,5)
    // found: mul(11,8)
    // found: do()
    // found: mul(8,5)
    // res += ( 8 * 5 = 40 ) = 48
    ```
    // This time, the sum of the results is **`48`** (`2*4 + 8*5`).
    - output: 48

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
    pattern = re.compile(r'''(?x)
                         mul \( ( \d{1,3} ) , ( \d{1,3} ) \) # mul(X,Y)
                         | ( do \( \) )                      # do()
                         | ( don't \( \) )                   # don't()
                         ''')

    enabled: bool = True
    res: int = 0
    for line in input:
        for match in pattern.finditer(line):
            if verbose:
                yield f'// found: {match.group(0)}'

            if match.group(3): # do()
                enabled = True

            elif match.group(4): # don't()
                enabled = False

            else: # mul(X,Y)
                if not enabled: continue

                sa = cast(str, match.group(1))
                sb = cast(str, match.group(2))
                a = int(sa)
                b = int(sb)
                m = a * b
                res += m
                if verbose > 1:
                    yield f'// res += ( {a} * {b} = {m} ) = {res}'

    yield f'{res}'

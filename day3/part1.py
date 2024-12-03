import re
import pytest
from collections.abc import Generator, Iterable
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    // TODO #given
    // TODO > input
    // TODO - verbose: 1
    // TODO - output: ```
    // TODO // how you got there
    // TODO what you got
    // TODO ```

''')), ids=MarkedSpec.get_id)
def test(spec: MarkedSpec):
    expected_output: list[str] = []
    verbose = False
    for name, value in spec.props:
        if name == 'output': expected_output.extend(spliterate(value, '\n'))
        elif name == 'verbose': verbose = any(value.lower().startswith(c) for c in 'ty')
        else: raise ValueError(f'invalid test prop {name!r}')
    lines = spliterate(spec.input, '\n')
    have_output = list(run(lines, verbose=1 if verbose else 0))
    assert have_output == expected_output

def run(input: Iterable[str], verbose: int = 0) -> Generator[str]:
    pattern = re.compile(r'(?x) # TODO what you want')

    for line in input:
        match = pattern.match(line)
        assert match

        if verbose:
            yield f'TODO parse {line!r}'

    yield f'TODO what you got'

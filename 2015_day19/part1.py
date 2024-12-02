import re
import pytest
from collections.abc import Generator, Iterable
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #given
    > FIXME
    - output: TODO

''')), ids=MarkedSpec.get_id)
def test(spec: MarkedSpec):
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

    for line in input:
        match = pattern.match(line)
        assert match

        raise NotImplementedError('TODO')

    yield f'TODO'

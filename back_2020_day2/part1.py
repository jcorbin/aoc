import re
from collections.abc import Generator, Iterable
from .strkit import spliterate, MarkedSpec

@MarkedSpec.mark('''

    // TODO #given
    // TODO > input
    // TODO - verbose: 1
    // TODO - output: result

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
    pattern = re.compile(r'(?x) # TODO input')

    for line in input:
        match = pattern.match(line)
        assert match

        if verbose > 1:
            yield f'// TODO how'

    yield f'TODO result'

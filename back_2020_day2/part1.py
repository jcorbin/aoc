import re
from collections.abc import Generator, Iterable
from typing import cast
from .strkit import spliterate, MarkedSpec

@MarkedSpec.mark('''

    // For example, suppose you have the following list:
    #given
    > 1-3 a: abcde
    > 1-3 b: cdefg
    > 2-9 c: ccccccccc
    // Each line gives the password policy and then the password.
    // The password policy indicates the lowest and highest number of times a given
    // letter must appear for the password to be valid.
    // For example, `1-3 a` means that the password must contain `a` at least `1` time
    // and at most `3` times.
    - verbose: 1
    - output: ```
    // invalid: cdefg have 0 b need 1-3
    2
    ```
    // In the above example, **`2`** passwords are valid.
    // The middle password, `cdefg`, is not; it contains no instances of `b`, but needs at least `1`.
    // The first and third passwords are valid: they contain one `a` or nine `c`,
    // both within the limits of their respective policies.

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
    pattern = re.compile(r'(?x) ( \d+ ) - ( \d+ ) \s+ ( \w ) : \s+ ( \w+ )')

    valid: int = 0
    for line in input:
        if not line.strip(): continue
        match = pattern.match(line)
        if not match:
            raise RuntimeError(f'invalid input {line!r}')
        slo = cast(str, match.group(1))
        shi = cast(str, match.group(2))
        let = cast(str, match.group(3))
        word = cast(str, match.group(4))

        lo = int(slo)
        hi = int(shi)
        n = word.count(let)
        if lo <= n <= hi:
            valid += 1
        elif verbose:
            yield f'// invalid: {word} have {n} {let} need {lo}-{hi}'

    yield f'{valid}'

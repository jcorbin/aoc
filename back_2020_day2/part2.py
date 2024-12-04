import re
from collections.abc import Generator, Iterable
from typing import cast
from .strkit import spliterate, MarkedSpec

@MarkedSpec.mark('''

    #given
    > 1-3 a: abcde
    > 1-3 b: cdefg
    > 2-9 c: ccccccccc
    // Each policy actually describes two **positions in the password**,
    // where `1` means the first character, `2` means the second character, and so on.
    // (Be careful; Toboggan Corporate Policies have no concept of "index zero"!)
    // **Exactly one of these positions** must contain the given letter.
    // Other occurrences of the letter are irrelevant for the purposes of policy enforcement.
    //
    // Given the same example list from above:
    //
    // - `1-3 a: abcde` is valid: position `1` contains `a` and position `3` does not.
    // - `1-3 b: cdefg` is invalid: neither position `1` nor position `3` contains `b`.
    // - `2-9 c: ccccccccc` is invalid: both position `2` and position `9` contain `c`.
    - verbose: 1
    - output: ```
    // invalid: cdefg have 0 b @1 or @3
    // invalid: ccccccccc have 2 c @2 or @9
    1
    ```

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

        n = 0
        if word[lo-1] == let: n += 1
        if word[hi-1] == let: n += 1

        if n == 1:
            valid += 1
        elif verbose:
            yield f'// invalid: {word} have {n} {let} @{lo} or @{hi}'

    yield f'{valid}'

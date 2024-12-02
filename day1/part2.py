import re
import pytest
from collections import Counter
from collections.abc import Generator, Iterable
from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #given
    > 3   4
    > 4   3
    > 2   5
    > 1   3
    > 3   9
    > 3   3
    // The first number in the left list is 3. It appears in the right list three times, so the similarity score increases by 3 * 3 = 9.
    // The second number in the left list is 4. It appears in the right list once, so the similarity score increases by 4 * 1 = 4.
    // The third number in the left list is 2. It does not appear in the right list, so the similarity score does not increase (2 * 0 = 0).
    // The fourth number, 1, also does not appear in the right list.
    // The fifth number, 3, appears in the right list three times; the similarity score increases by 9.
    // The last number, 3, appears in the right list three times; the similarity score again increases by 9.
    // So, for these example lists, the similarity score at the end of this process is 31 (9 + 4 + 0 + 0 + 9 + 9).
    - output: 31

''')))
def test(spec: MarkedSpec):
    lines = spliterate(spec.input, '\n')
    expected_output: list[str] = []

    for name, value in spec.props:
        if name == 'output': expected_output.append(value)
        else: raise ValueError(f'invalid test prop {name!r}')

    assert list(run(lines)) == expected_output

def run(input: Iterable[str]) -> Generator[str]:
    pattern = re.compile(r'(?x) ( \d+ ) \s+ ( \d+ )')

    a_list: list[int] = []
    b_list: list[int] = []

    for line in input:
        match = pattern.match(line)
        assert match
        sa, sb = match.groups()
        a = int(sa)
        b = int(sb)
        a_list.append(a)
        b_list.append(b)

    b_freq = Counter(b_list)
    res: int = 0
    for a in a_list:
        res += a * b_freq.get(a, 0)

    yield f'{res}'

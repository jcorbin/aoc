import re
import pytest
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
    // The smallest number in the left list is 1, and the smallest number in the right list is 3. The distance between them is 2.
    // The second-smallest number in the left list is 2, and the second-smallest number in the right list is another 3. The distance between them is 1.
    // The third-smallest number in both lists is 3, so the distance between them is 0.
    // The next numbers to pair up are 3 and 4, a distance of 1.
    // The fifth-smallest numbers in each list are 3 and 5, a distance of 2.
    // Finally, the largest number in the left list is 4, while the largest number in the right list is 9; these are a distance 5 apart.
    // In the example above, this is 2 + 1 + 0 + 1 + 2 + 5, a total distance of 11!
    - output: 11

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

    a_list = sorted(a_list)
    b_list = sorted(b_list)
    res: int = 0
    for a, b in zip(a_list, b_list):
        dab = abs(a - b)
        res += dab

    yield f'{res}'

def main():
    import sys
    for line in run(sys.stdin):
        print(line)

if __name__ == '__main__':
    main()

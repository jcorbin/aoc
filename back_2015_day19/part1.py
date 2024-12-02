import pytest
import re
from collections import defaultdict
from collections.abc import Generator, Iterable
from typing import cast, final

from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #HOH
    > H => HO
    > H => OH
    > O => HH
    >
    > HOH
    // Given the replacements above and starting with `HOH`, the following
    // molecules could be generated:
    //
    // - `HOOH` (via `H => HO` on the first `H`).
    // - `HOHO` (via `H => HO` on the second `H`).
    // - `OHOH` (via `H => OH` on the first `H`).
    // - `HOOH` (via `H => OH` on the second `H`).
    // - `HHHH` (via `O => HH`).
    //
    // So, in the example above, there are `4` **distinct** molecules (not five,
    // because `HOOH` appears twice) after one replacement from `HOH`.
    - verbose: yes
    - output: ```
    // HOOH
    // OHOH
    // HHHH
    // HOHO
    // HOOH (dupe)
    4
    ```

    // Santa's favorite molecule, `HOHOHO`, can become `7` **distinct**
    // molecules (over nine replacements: six from `H`, and three from `O`).
    #HOHOHO
    > H => HO
    > H => OH
    > O => HH
    >
    > HOHOHO
    - output: 7

    // The machine replaces without regard for the surrounding characters.
    // For example, given the string `H2O`, the transition `H => OO` would result in `OO2O`.
    #H20
    > H => OO
    > 
    > H2O
    - verbose: yes
    - output: ```
    // OO2O
    1
    ```

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

@final
class Machine:
    def __init__(self, rules: Iterable[tuple[str, Iterable[str]]]):
        self.xlate = {a: tuple(bs) for a, bs in rules}
        self.xlate_pattern = re.compile('|'.join(self.xlate.keys()))

    def examine(self, s: str):
        parts: list[str] = []
        refs: list[str] = []
        k = 0
        for match in self.xlate_pattern.finditer(s):
            xk = match.group(0)
            i = match.start(0)
            parts.append(s[k:i])
            refs.append(xk)
            k = match.end(0)
        parts.append(s[k:])
        return parts, refs

    def expand1(self, parts: list[str], refs: list[str]) -> Generator[str]:
        reps = [self.xlate[ref] for ref in refs]

        prefix = ''
        for i, ref in enumerate(refs):
            prefix += parts[i]

            suffix = ''
            for j in range(i+1, len(parts)):
                suffix += parts[j]
                if j < len(refs): suffix += refs[j]

            may_rep = reps[i]
            for rep in may_rep:
                yield f'{prefix}{rep}{suffix}'

            prefix += ref

def run(input: Iterable[str], verbose: bool=False) -> Generator[str]:
    rule = re.compile(r'(?x) ( \w+ ) \s* => \s* ( \w+ ) $')

    rules: defaultdict[str, list[str]] = defaultdict(list)
    for line in input:
        if not line.strip(): break
        match = rule.match(line)
        if not match: raise RuntimeError(f'invalid input {line!r}')
        pat = cast(str, match.group(1))
        rep = cast(str, match.group(2))
        rules[pat].append(rep)

    mach = Machine(rules.items())

    for line in input:
        line = line.strip()
        parts, refs = mach.examine(line)

        # yield f'// parts: {parts!r}'
        # yield f'// refs: {refs!r}'

        seen: set[str] = set()
        if verbose:
            for sub in mach.expand1(parts, refs):
                if sub in seen:
                    yield f'// {sub} (dupe)'
                else:
                    yield f'// {sub}'
                    seen.add(sub)
        else:
            seen.update(mach.expand1(parts, refs))

        yield f'{len(seen)}'

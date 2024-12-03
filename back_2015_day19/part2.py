import heapq
import pytest
import re
from collections import defaultdict
from collections.abc import Generator, Iterable
from typing import cast, final

from .strkit import spliterate, MarkedSpec

@pytest.mark.parametrize('spec', list(MarkedSpec.iterspecs('''

    #HOH
    > e => H
    > e => O
    > H => HO
    > H => OH
    > O => HH
    >
    > HOH
    // If you'd like to make `HOH`, you start with `e`,
    // and then make the following replacements:
    //
    // - `e => O` to get `O`
    // - `O => HH` to get `HH`
    // - `H => OH` (on the second `H`) to get `HOH`
    //
    // So, you could make `HOH` after **`3` steps**.
    - verbose: 0
    - output: ```
    3
    ```

    // Santa's favorite molecule, `HOHOHO`, can be made in **`6` steps**.
    #HOHOHO
    > e => H
    > e => O
    > H => HO
    > H => OH
    > O => HH
    >
    > HOHOHO
    - verbose: 0
    - output: ```
    6
    ```

''')), ids=MarkedSpec.get_id)
def test(spec: MarkedSpec):
    lines = spliterate(spec.input, '\n')
    expected_output: list[str] = []
    verbose = 0
    for name, value in spec.props:
        if name == 'output': expected_output.extend(spliterate(value, '\n'))
        elif name == 'verbose': verbose = int(value)
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

        # TODO regex too slow below, but this dont work yet
        # KN = self.keysize
        # i = 0
        # j = i
        # while j < len(s):
        #     for n in range(1, KN+1):
        #         k = j + n
        #         xk = s[j:k]
        #         if xk in self.xlate:
        #             parts.append(s[i:j])
        #             refs.append(xk)
        #             i = j = k
        #             break
        #     else:
        #         j += 1
        # parts.append(s[i:])

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

'''
From Wikipedia article; Iterative with two matrix rows.

Modified slightly form
<https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#Python>
'''
def levenshtein(s: str, t: str):
    if s == t: return 0
    N = len(s)
    M = len(t)
    if N == 0: return M
    if M == 0: return N
    v0 = list(range(M+1))
    v1: list[int] = [0] * (M + 1)
    for i in range(len(s)):
        v1[0] = i + 1
        for j in range(len(t)):
            cost = 0 if s[i] == t[j] else 1
            v1[j + 1] = min(v1[j] + 1, v0[j + 1] + 1, v0[j] + cost)
        for j in range(len(v0)):
            v0[j] = v1[j]
    return v1[len(t)]

'''
Further modified above for 1 fixed point
'''
def vantagestein(v: str):
    M = len(v)
    v0: list[int] = [0] * (M + 1) # list(range(M+1))
    v1: list[int] = [0] * (M + 1)

    def dist(s: str):
        if s == v: return 0
        N = len(s)
        if N == 0: return M
        if M == 0: return N

        for i in range(len(v0)): v0[i] = i
        for i in range(len(s)):
            v1[0] = i + 1
            for j in range(len(v)):
                cost = 0 if s[i] == v[j] else 1
                v1[j + 1] = min(v1[j] + 1, v0[j + 1] + 1, v0[j] + cost)
            for j in range(len(v0)):
                v0[j] = v1[j]
        return v1[len(v)]

    return dist

@final
class Search:
    State = tuple[int, int, str]

    def __init__(self, mach: Machine, start: str, goal: str):
        self.mach = mach
        self.goal = goal
        self.seen: set[str] = set()
        self.space: list[Search.State] = []
        self.dist = vantagestein(self.goal)

        # dist = levenshtein(self.goal, start)
        dist = self.dist(start)
        heapq.heappush(self.space, (dist, 0, start))

    def expand_all(self):
        space: list[Search.State] = []
        res: int|None = None

        for d, steps, s in self.space:
            nsteps = steps+1

            nss = set(self.mach.expand1(*self.mach.examine(s)))
            if self.goal in nss:
                if res is None or nsteps < res:
                    res = nsteps
                continue

            nss.difference_update(self.seen)
            self.seen.update(nss)

            nss = [ns for ns in nss if len(ns) <= len(self.goal)]

            # nds = [levenshtein(self.goal, ns) for ns in nss]
            nds = [self.dist(ns) for ns in nss]

            space.extend(
                (nd, nsteps, ns)
                for nd, ns in zip(nds, nss)
                if nd <= d
            )

        # heapq.heapify(space)

        self.space = space

        return res

    def expand(self):
        heapq.heapify(self.space)

        d, steps, s = heapq.heappop(self.space)
        nsteps = steps+1

        nss = set(self.mach.expand1(*self.mach.examine(s)))
        if self.goal in nss: return nsteps

        nss.difference_update(self.seen)
        self.seen.update(nss)

        # TODO seems like a good idea, but fails a test case
        # self.space = [st for st in self.space if st[2] not in self.seen]

        nss = [ns for ns in nss if len(ns) <= len(self.goal)]

        # nds = [levenshtein(self.goal, ns) for ns in nss]
        nds = [self.dist(ns) for ns in nss]

        nst = [(nd, nsteps, ns) for nd, ns in zip(nds, nss) if nd <= d]

        self.space.extend(nst)

        return None

def run(input: Iterable[str], verbose: int = 0) -> Generator[str]:
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
        search = Search(mach, 'e', line)
        best: int = 0
        rounds: int = 0
        
        while search.space:
            rounds += 1
            if verbose:
                yield f'// rounds:{rounds} {search.space[0][0]} #{len(search.space)}'

            # steps = search.expand_all()
            # if steps is not None:
            #     if not best or steps < best:
            #         best = steps
            #         if verbose:
            #             yield f'// found #{steps}'

            steps = search.expand()
            if steps is not None:
                if not best or steps < best:
                    best = steps
                    if verbose:
                        yield f'// found #{steps}'

        yield f'{best}'

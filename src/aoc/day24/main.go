package main

import (
	"aoc/internal/infernio"
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jcorbin/anansi"
)

var (
	boost      = flag.Int("boost", 0, "boost value")
	boostGroup = flag.String("boostGroup", "Immune System", "boost group")
	verbose    = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

type scenario struct {
	groupOrder []string
	groups     map[string][]group
}

type group struct {
	n      int
	hp     int
	weak   map[string]struct{}
	immune map[string]struct{}
	damage int
	attack string
	init   int
}

var builtinInput = infernio.Builtin("" +
	"Immune System:\n" +
	"17 units each with 5390 hit points (weak to radiation, bludgeoning) with an attack that does 4507 fire damage at initiative 2\n" +
	"989 units each with 1274 hit points (immune to fire; weak to bludgeoning, slashing) with an attack that does 25 slashing damage at initiative 3\n" +
	"\n" +
	"Infection:\n" +
	"801 units each with 4706 hit points (weak to radiation) with an attack that does 116 bludgeoning damage at initiative 1\n" +
	"4485 units each with 2961 hit points (immune to radiation; weak to fire, cold) with an attack that does 12 slashing damage at initiative 4\n")

func run(in, out *os.File) error {
	var verboseOut io.Writer
	if *verbose {
		verboseOut = out
	}

	var scene scenario

	runRound := func(bg string, b int) bool {
		if b != 0 {
			scene.applyBoost(bg, b)
		}
		scene.run(verboseOut)
		winningName, winningScore := scene.winningScore()
		if winningScore == 0 {
			log.Printf("stasis")
			return false
		}
		log.Printf("%s won with %v", winningName, winningScore)
		return winningName == *boostGroup
	}

	if err := infernio.LoadInput(builtinInput, scene.load); err != nil {
		return err
	}

	if *boost != 0 {
		runRound(*boostGroup, *boost)
		return nil
	}

	log.Printf("searching for min boost for %q", *boostGroup)
	for b := 0; ; b++ {
		orig := scene.copy()
		if runRound(*boostGroup, b) {
			log.Printf("minimum boost: %v", b)
			break
		}
		scene = orig.copy()
	}

	return nil
}

func (scene *scenario) copy() scenario {
	var c scenario
	c.groups = make(map[string][]group, len(scene.groups))
	for groupName, gs := range scene.groups {
		c.groups[groupName] = append([]group(nil), gs...)
	}
	c.groupOrder = append([]string(nil), scene.groupOrder...)
	return c
}

func (scene *scenario) run(w io.Writer) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	n := 0
	t0 := time.Now()
	for len(scene.groups) > 1 {
		select {
		case <-tick.C:
			t1 := time.Now()
			log.Printf("running... %v in %v (%.1f/s)",
				n, t1.Sub(t0),
				float64(n)/(float64(t1.Sub(t0))/float64(time.Second)),
			)
			// if w == nil {
			// 	w = os.Stdout
			// }
		default:
		}
		n++
		if w != nil {
			scene.survey(w)
			io.WriteString(w, "\n")
		}
		if w != nil {
			scene.rankGroups(w)
			io.WriteString(w, "\n")
		}
		if scene.attack(w, scene.selectTargets()) == 0 {
			if w != nil {
				io.WriteString(w, "no winner, stasis reached\n")
			}
			return
		}
		scene.cleanup()
		if w != nil {
			io.WriteString(w, "\n")
		}
		// After the fight is over, if both armies still contain units, a new fight
		// begins; combat only ends once one army has lost all of its units.
	}
	if w != nil {
		scene.survey(w)
	}
}

func (scene *scenario) applyBoost(groupName string, dmg int) {
	gs := scene.groups[groupName]
	for i := range gs {
		gs[i].damage += dmg
	}
}

func (scene *scenario) winningScore() (string, int) {
	for len(scene.groups) == 1 {
		for _, groupName := range scene.groupOrder {
			gs := scene.groups[groupName]
			if len(gs) == 0 {
				continue
			}
			n := 0
			for _, g := range gs {
				n += g.n
			}
			return groupName, n
		}
	}
	return "", 0
}

func (scene *scenario) survey(w io.Writer) {
	for _, groupName := range scene.groupOrder {
		fmt.Fprintf(w, "%s:\n", groupName)
		if gs := scene.groups[groupName]; len(gs) == 0 {
			fmt.Fprintf(w, "No groups remain.\n")
		} else {
			for i, g := range gs {
				fmt.Fprintf(w, "Group %v contains %v units\n", i+1, g.n)
			}
		}
	}
}

func (g group) effectivePower() int { return g.n * g.damage }

type groupID struct {
	name string
	i    int
}

func (scene *scenario) cleanup() {
	for groupName, gs := range scene.groups {
		i := 0
		for j := 0; j < len(gs); j++ {
			if gs[j].n <= 0 {
				continue
			}
			if i != j {
				gs[i] = gs[j]
			}
			i++
		}
		if i > 0 {
			scene.groups[groupName] = gs[:i]
		} else {
			delete(scene.groups, groupName)
		}
	}
}

func (scene *scenario) attack(w io.Writer, targets map[groupID]groupID) (totalKilled int) {
	// During the *attacking* phase, each group deals damage to the target it
	// selected, if any. Groups attack in decreasing order of initiative,
	// regardless of whether they are part of the infection or the immune
	// system. (If a group contains no units, it cannot attack.)
	order := make([]groupID, 0, len(targets))
	for id := range targets {
		order = append(order, id)
	}
	sort.Slice(order, func(i, j int) bool {
		a := scene.groups[order[i].name][order[i].i]
		b := scene.groups[order[j].name][order[j].i]
		return a.init > b.init
	})

	// The damage an attacking group deals to a defending group depends on the
	// attacking group\'s attack type and the defending group\'s immunities and
	// weaknesses. By default, an attacking group would deal damage equal to
	// its *effective power* to the defending group. However, if the defending
	// group is *immune* to the attacking group\'s attack type, the defending
	// group instead takes *no damage*; if the defending group is *weak* to the
	// attacking group\'s attack type, the defending group instead takes
	// *double damage*.
	// dead := make(map[groupID]struct{}, len(targets))
	for _, atkID := range order {
		// The defending group only loses *whole units* from damage; damage is
		// always dealt in such a way that it kills the most units possible,
		// and any remaining damage to a unit that does not immediately kill it
		// is ignored. For example, if a defending group contains `10` units
		// with `10` hit points each and receives `75` damage, it loses exactly
		// `7` units and is left with `3` units at full health.
		defID := targets[atkID]

		atk := scene.groups[atkID.name][atkID.i]
		if atk.n <= 0 {
			continue
		}

		def := scene.groups[defID.name][defID.i]
		dmg := atk.effectiveDamage(def)
		killed := dmg / def.hp
		if killed > def.n {
			killed = def.n
		}
		def.n -= killed
		if w != nil {
			fmt.Fprintf(w,
				"%v group %v attacks defending group %v, killing %v units\n",
				atkID.name, atkID.i+1, defID.i+1, killed)
		}
		totalKilled += killed
		scene.groups[defID.name][defID.i] = def
	}
	return totalKilled
}

func (scene *scenario) selectTargets() map[groupID]groupID {
	// During the *target selection* phase, each group attempts to choose one
	// target.

	// In decreasing order of effective power, groups choose their targets; in
	// a tie, the group with the higher initiative chooses first.
	var order []groupID
	for _, groupName := range scene.groupOrder {
		for i := range scene.groups[groupName] {
			order = append(order, groupID{groupName, i})
		}
	}
	sort.Slice(order, func(i, j int) bool {
		a := scene.groups[order[i].name][order[i].i]
		b := scene.groups[order[j].name][order[j].i]
		if a.effectivePower() > b.effectivePower() {
			return true
		}
		return a.effectivePower() == b.effectivePower() && a.init > b.init
	})

	targets := make(map[groupID]groupID, len(order))
	attacker := make(map[groupID]groupID, len(order))

	for _, atkID := range order {
		// The attacking group chooses to target the group in the enemy army to
		// which it would deal the most damage (after accounting for weaknesses and
		// immunities, but not accounting for whether the defending group has
		// enough units to actually receive all of that damage).
		atk := scene.groups[atkID.name][atkID.i]
		var targetID groupID
		var targ group
		var best int
		for _, id := range order {
			if id.name == atkID.name {
				continue // same side
			}
			if _, taken := attacker[id]; taken {
				continue
			}
			def := scene.groups[id.name][id.i]
			dmg := atk.effectiveDamage(def)
			if dmg > best {
				targetID, targ, best = id, def, dmg
			} else if dmg == best {
				// If an attacking group is considering two defending groups to
				// which it would deal equal damage, it chooses to target the
				// defending group with the largest effective power; if there
				// is still a tie, it chooses the defending group with the
				// highest initiative. If it cannot deal any defending groups
				// damage, it does not choose a target. Defending groups can
				// only be chosen as a target by one attacking group.
				if ap, bp := def.effectivePower(), targ.effectivePower(); ap > bp || (ap == bp && def.init > targ.init) {
					// log.Printf(
					// 	"tie++ def %v/%v => targ %v/%v",
					// 	def.effectivePower(), def.init,
					// 	targ.effectivePower(), targ.init,
					// )
					targetID, targ, best = id, def, dmg
				}
			}
			// log.Printf("%v => %v ? %v => %v", atkID, id, dmg, targetID)
		}
		if best > 0 {
			targets[atkID] = targetID
			attacker[targetID] = atkID
		}
	}

	return targets
}

func (g group) effectiveDamage(def group) int {
	if _, immune := def.immune[g.attack]; immune {
		return 0
	}
	dmg := g.effectivePower()
	if _, weak := def.weak[g.attack]; weak {
		dmg *= 2
	}
	return dmg
}

func (scene *scenario) rankGroups(w io.Writer) {
	for ai, groupName := range scene.groupOrder {
		for i, atk := range scene.groups[groupName] {
			for di, defGroupName := range scene.groupOrder {
				if di == ai {
					continue
				}
				n := 0
				for j, def := range scene.groups[defGroupName] {
					dmg := atk.effectiveDamage(def)
					if dmg > 0 {
						fmt.Fprintf(w,
							"%s group %v would deal defending group %v %v damage\n",
							groupName,
							i+1,
							j+n+1,
							dmg,
						)
					}
				}
				n += len(scene.groups[defGroupName])
			}
		}
	}
}

var (
	headerPattern = regexp.MustCompile(`^(.+):$`)
	attrPattern   = regexp.MustCompile(`(weak|immune) to (\w+(?:, \w+)*)`)
	groupPattern  = regexp.MustCompile(`^(\d+) units each with (\d+) hit points(?: \((.+?)\))? with an attack that does (\d+) (\w+) damage at initiative (\d+)$`)
)

func (scene *scenario) load(r io.Reader) error {
	sc := bufio.NewScanner(r)

	expect := func(pat *regexp.Regexp) ([]string, error) {
		parts := pat.FindStringSubmatch(sc.Text())
		if len(parts) == 0 {
			return nil, fmt.Errorf("bad line %q, expecting %v", sc.Text(), pat)
		}
		return parts, nil
	}

	scene.groups = make(map[string][]group, 2)

	var groupName string
	for sc.Scan() {
		line := sc.Text()

		// group separator
		if line == "" {
			groupName = ""
			continue
		}

		// expect header
		if groupName == "" {
			parts, err := expect(headerPattern)
			if err != nil {
				return err
			}
			if _, def := scene.groups[parts[1]]; def {
				return fmt.Errorf("duplicate group %q", parts[1])
			}
			groupName = parts[1]
			scene.groupOrder = append([]string{groupName}, scene.groupOrder...)
			continue
		}

		// expect a group
		parts, err := expect(groupPattern)
		if err != nil {
			return err
		}

		var g group
		g.n, _ = strconv.Atoi(parts[1])
		g.hp, _ = strconv.Atoi(parts[2])

		for _, attrParts := range attrPattern.FindAllStringSubmatch(parts[3], -1) {
			switch attrParts[1] {
			case "weak":
				if g.weak == nil {
					g.weak = make(map[string]struct{})
				}
				for _, attack := range strings.Split(attrParts[2], ", ") {
					g.weak[attack] = struct{}{}
				}
			case "immune":
				if g.immune == nil {
					g.immune = make(map[string]struct{})
				}
				for _, attack := range strings.Split(attrParts[2], ", ") {
					g.immune[attack] = struct{}{}
				}
			default:
				panic("inconceivable")
			}
		}

		g.damage, _ = strconv.Atoi(parts[4])
		g.attack = parts[5]
		g.init, _ = strconv.Atoi(parts[6])

		scene.groups[groupName] = append(scene.groups[groupName], g)
	}
	return sc.Err()
}

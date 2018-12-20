package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"aoc/internal/geom"
	"aoc/internal/infernio"
	"aoc/internal/layerui"
	"aoc/internal/quadindex"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	logfile  = flag.String("logfile", "", "log file")
	verbose  = flag.Bool("v", false, "verbose logs")
	goblinAP = flag.Int("goblin-ap", 3, "goblin attack power")
	elfinAP  = flag.Int("elf-ap", 3, "elf attack power")
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		// for _, s := range []string{
		// } {
		// 	s = strings.Replace(s, "_", " ", -1)
		// 	io.WriteString(out, s)
		// }
		fmt.Fprintf(out, "\n\nUsage %s [options] [<inputFile>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	anansi.MustRun(layerui.WithOpenLogFile(*logfile, run))
}

func run() error {
	var world gameWorld
	if err := infernio.LoadInput(builtinInput, world.load); err != nil {
		return err
	}

	foc := world.bounds.Size().Div(2)
	foc.X /= 2
	foc.X *= 3
	worldLayer := layerui.WorldLayer{World: &world}
	worldLayer.SetFocus(foc)

	return layerui.Layers(
		&layerui.LogLayer{SubGrid: layerui.BottomNLines(5)},
		&worldLayer,
	).RunMain()
}

var builtinInput = infernio.Builtin("" +
	"#########\n" +
	"#G......#\n" +
	"#.E.#...#\n" +
	"#..##..G#\n" +
	"#...##..#\n" +
	"#...#...#\n" +
	"#.G...G.#\n" +
	"#.....G.#\n" +
	"#########")

type gameType uint8

const (
	gameRender gameType = 1 << iota
	gameCollide
	gameActor
	gameHP
	gameAP
	gameLabel
)

type gameWorld struct {
	sidebar   bytes.Buffer
	needsDraw time.Duration
	bounds    image.Rectangle

	quadindex.Index

	round int

	t  []gameType
	p  []image.Point
	z  []int
	r  []byte
	a  []ansi.SGRAttr
	hp []int
	ap []int

	free     []int
	noTarget bool
	actors   []int
	goblins  map[int]struct{}
	elves    map[int]struct{}
	deaths   int
}

func (world *gameWorld) Bounds() image.Rectangle {
	return world.bounds
}

func (world *gameWorld) NeedsDraw() time.Duration {
	return world.needsDraw
}

func (world *gameWorld) describe(id int, buf *bytes.Buffer) {
	t := world.t[id]
	fmt.Fprintf(buf, "id:%v t:%02x", id, t)
	if t&gameRender != 0 {
		fmt.Fprintf(buf, " p:%v z:%v r:%q a:%v", world.p[id], world.z[id], world.r[id], world.a[id])
	}
	if t&gameCollide != 0 {
		buf.WriteString(" collides")
	}
	if t&gameHP != 0 {
		fmt.Fprintf(buf, " hp:%v", world.hp[id])
	}
	if t&gameAP != 0 {
		fmt.Fprintf(buf, " ap:%v", world.ap[id])
	}
	buf.WriteRune('\n')
}

func (world *gameWorld) Render(g anansi.Grid, viewOffset image.Point) {
	// draw primary grid
	gz := make([]int, len(g.Rune)) // TODO re-use allocation
	for id, t := range world.t {
		if id == 0 || t&gameRender == 0 {
			continue
		}
		sp := world.p[id].Add(viewOffset)
		if sp.X < 0 || sp.Y < 0 {
			continue
		}
		gi, ok := g.CellOffset(ansi.PtFromImage(sp))
		if !ok {
			continue
		}
		if z := world.z[id]; z > gz[gi] {
			gz[gi] = z
			g.Rune[gi] = rune(world.r[id])
			g.Attr[gi] = mergeBGColors(world.a[id], g.Attr[gi])
		} else {
			g.Attr[gi] = mergeBGColors(g.Attr[gi], world.a[id])
		}
	}

	// update hp sidebar

	// TODO switch to left side if more space
	sidebarAt := image.Pt(
		world.bounds.Max.X+2, // TODO why 2?
		world.bounds.Min.Y,
	).Add(viewOffset)
	if sidebarAt.Y < 0 {
		sidebarAt.Y = 0
	}
	if sidebarAt.X < 0 {
		return
	}
	sidebarGrid := g.SubAt(ansi.PtFromImage(sidebarAt))

	world.sidebar.Reset()
	world.sortIDs(world.actors)
	bnd := g.Bounds()
	cur := sidebarGrid.Rect.Min
	cur.X = sidebarGrid.Rect.Min.X
	for _, id := range world.actors {
		sp := world.p[id].Add(viewOffset)
		if sp.X < 0 || sp.Y < 0 {
			continue
		}
		gp := ansi.PtFromImage(sp)
		if !gp.In(bnd) {
			continue
		}

		for ; cur.Y < gp.Y; cur.Y++ {
			world.sidebar.WriteByte('\n')
			cur.X = sidebarGrid.Rect.Min.X
		}

		if cur.X > sidebarGrid.Rect.Min.X {
			world.sidebar.WriteByte(' ')
			cur.X++
		}

		t := world.t[id]
		if t&gameRender != 0 {
			world.sidebar.WriteString(world.a[id].ControlString())
			world.sidebar.WriteByte(world.r[id])
			world.sidebar.WriteString("\x1b[0m")
			cur.X++
		}
		if t&gameHP != 0 {
			world.sidebar.WriteByte('(')
			n, _ := world.sidebar.WriteString(strconv.Itoa(world.hp[id]))
			world.sidebar.WriteByte(')')
			cur.X += 1 + n + 1
		}
		// TODO optional #id annotation
		// world.sidebar.WriteByte('#')
		// n, _ := world.sidebar.WriteString(strconv.Itoa(id))
		// cur.X += 1 + n

	}

	layerui.WriteIntoGrid(sidebarGrid, world.sidebar.Bytes())
}

func mergeBGColors(a, b ansi.SGRAttr) ansi.SGRAttr {
	if _, hasBG := a.BG(); !hasBG {
		if c, def := b.BG(); def {
			a |= c.BG()
		}
	}
	return a
}

func (world *gameWorld) load(r io.Reader) error {
	if len(world.t) > 0 {
		panic("reload of world not supported")
	}
	world.createEntity(0)

	// scan entities from input
	if nom, ok := r.(interface{ Name() string }); ok {
		log.Printf("read input from %s", nom.Name())
	}
	sc := bufio.NewScanner(r)
	var p image.Point
	for sc.Scan() {
		line := sc.Text()
		p.X = 0
		for i := 0; i < len(line); i++ {
			world.loadCell(p, line[i])
			p.X++
		}
		p.Y++
	}

	world.bounds = world.computeBounds()

	world.actors = make([]int, 0, len(world.t))
	for id, t := range world.t {
		if t&gameActor != 0 {
			world.actors = append(world.actors, id)
		}
	}
	world.goblins = make(map[int]struct{}, len(world.actors))
	world.elves = make(map[int]struct{}, len(world.actors))
	for _, id := range world.actors {
		switch world.r[id] {
		case 'G':
			world.goblins[id] = struct{}{}
		case 'E':
			world.elves[id] = struct{}{}
		}
	}
	return sc.Err()
}

func (world *gameWorld) loadCell(p image.Point, c byte) {
	switch c {
	case '#':
		world.createWall(p)
	case '.':
		world.createFloor(p)
	case 'G':
		world.createActor(p, 'G', ansi.RGB(128, 32, 16), 200, *goblinAP)
		world.createFloor(p)
	case 'E':
		world.createActor(p, 'E', ansi.RGB(16, 128, 32), 200, *elfinAP)
		world.createFloor(p)
	}
}

func (world *gameWorld) createWall(p image.Point) int {
	id := world.createEntity(gameRender | gameCollide)
	world.p[id] = p
	world.z[id] = 2
	world.r[id] = '#'
	world.a[id] = ansi.RGB(128, 128, 128).FG() | ansi.RGB(32, 32, 32).BG()
	world.Index.Update(id, p)
	return id
}

func (world *gameWorld) createFloor(p image.Point) int {
	id := world.createEntity(gameRender)
	world.p[id] = p
	world.z[id] = 1
	world.r[id] = '.'
	world.a[id] = ansi.RGB(32, 32, 32).FG() | ansi.RGB(16, 16, 16).BG()
	world.Index.Update(id, p)
	return id
}

func (world *gameWorld) createActor(p image.Point, r byte, c ansi.SGRColor, hp, ap int) int {
	id := world.createEntity(gameRender | gameCollide | gameActor | gameHP | gameAP)
	world.p[id] = p
	world.z[id] = 10
	world.r[id] = r
	world.a[id] = c.FG()
	world.hp[id] = hp
	world.ap[id] = ap
	world.Index.Update(id, p)
	return id
}

func (world *gameWorld) createEntity(t gameType) int {
	if len(world.free) > 0 {
		id := world.free[len(world.free)-1]
		world.free = world.free[:len(world.free)-1]
		return id
	}
	id := len(world.t)
	world.t = append(world.t, t)
	world.p = append(world.p, image.ZP)
	world.z = append(world.z, 0)
	world.r = append(world.r, 0)
	world.a = append(world.a, 0)
	world.hp = append(world.hp, 0)
	world.ap = append(world.ap, 0)
	return id
}

func (world *gameWorld) pruneActors() {
	i := 0
	for j := 0; j < len(world.actors); j++ {
		if world.actors[j] != 0 {
			world.actors[i] = world.actors[j]
			i++
		}
	}
	world.actors = world.actors[:i]
}

func (world *gameWorld) destroyEntity(id int) {
	if _, weCare := world.elves[id]; weCare {
		// no one counts goblin deaths
		world.deaths++
	}

	for i := 0; i < len(world.actors); i++ {
		if world.actors[i] == id {
			world.actors[i] = 0
		}
	}
	delete(world.goblins, id)
	delete(world.elves, id)
	world.Index.Delete(id, world.p[id])
	world.t[id] = 0
	world.p[id] = image.ZP
	world.z[id] = 0
	world.r[id] = 0
	world.a[id] = 0
	world.hp[id] = 0
	world.ap[id] = 0
	world.free = append(world.free, id)
}

func (world *gameWorld) computeBounds() (bounds image.Rectangle) {
	if len(world.t) == 0 {
		return bounds
	}
	id := 1
	for ; id < len(world.t); id++ {
		if world.t[id]&gameRender != 0 {
			bounds.Min = world.p[id]
			bounds.Max = world.p[id]
			break
		}
	}
	for ; id < len(world.t); id++ {
		if world.t[id]&gameRender == 0 {
			continue
		}
		p := world.p[id]
		if bounds.Min.X > p.X {
			bounds.Min.X = p.X
		}
		if bounds.Min.Y > p.Y {
			bounds.Min.Y = p.Y
		}
		if bounds.Max.X < p.X {
			bounds.Max.X = p.X
		}
		if bounds.Max.Y < p.Y {
			bounds.Max.Y = p.Y
		}
	}
	return bounds
}

func (world *gameWorld) done() bool {
	if world.noTarget {
		return true
	}
	return false
}

func (world *gameWorld) winningTeam() string {
	if len(world.goblins) == 0 {
		return "elves"
	}
	if len(world.elves) == 0 {
		return "goblins"
	}
	return "none"
}

func (world *gameWorld) remainingHP() (hp int) {
	for _, actorID := range world.actors {
		hp += world.hp[actorID]
	}
	return hp
}

func (world *gameWorld) finish() {
	log.Printf("finished after round %v: %s won with %v remaining hp",
		world.round,
		world.winningTeam(),
		world.remainingHP(),
	)
	if world.deaths > 0 {
		log.Printf("\x1b[91mFAIL\x1b[0m there were %v elf deaths...", world.deaths)
	} else {
		log.Printf("\x1b[92mSUCCESS\x1b[0m no elf deaths!")
		log.Printf("outcome is %v", world.round*world.remainingHP())
	}
}

func (world *gameWorld) Tick() bool {
	if world.done() {
		return false
	}

	world.sortIDs(world.actors)
	for _, id := range world.actors {
		if !world.act(id) {
			if world.done() {
				world.finish()
				return false
			}
		}
	}
	world.pruneActors()
	world.round++

	return true
}

func (world *gameWorld) act(actorID int) bool {
	// already killed by prior actor in this round
	if actorID == 0 {
		return true
	}

	var enemySet map[int]struct{}
	if _, isGoblin := world.goblins[actorID]; isGoblin {
		enemySet = world.elves
	} else if _, isElf := world.elves[actorID]; isElf {
		enemySet = world.goblins
	} else {
		panic(fmt.Sprintf("neither goblin nor elf #%v", actorID))
	}

	// done if no enemies left
	if len(enemySet) == 0 {
		world.noTarget = true
		return false
	}

	attackID := world.chooseAttack(actorID, enemySet)
	if attackID == 0 {
		if !world.move(actorID, enemySet) {
			return true
		}
		attackID = world.chooseAttack(actorID, enemySet)
	}

	if attackID != 0 {
		world.attack(actorID, attackID)
	}

	return true
}

func (world *gameWorld) chooseAttack(actorID int, enemySet map[int]struct{}) int {
	// weakest adjacent enemy id
	actorP := world.p[actorID]
	attackID, attackHP := 0, 0
	for _, ep := range [4]image.Point{
		actorP.Add(image.Pt(0, -1)),
		actorP.Add(image.Pt(-1, 0)),
		actorP.Add(image.Pt(1, 0)),
		actorP.Add(image.Pt(0, 1)),
	} {
		if enemyID := world.enemyAt(ep, enemySet); enemyID != 0 {
			if attackID == 0 || world.hp[enemyID] < attackHP {
				attackID = enemyID
				attackHP = world.hp[enemyID]
			}
		}
	}
	return attackID
}

func (world *gameWorld) attack(actorID, attackID int) {
	if *verbose {
		log.Printf(
			"attack %s@%v#%v => %s@%v#%v",
			string(world.r[actorID]), world.p[actorID], actorID,
			string(world.r[attackID]), world.p[attackID], attackID,
		)
	}
	if hp := world.hp[attackID] - world.ap[actorID]; hp <= 0 {
		world.destroyEntity(attackID)
	} else {
		world.hp[attackID] = hp
	}
}

func (world *gameWorld) move(actorID int, enemySet map[int]struct{}) bool {
	// TODO re-use
	var reach reachabilityScore
	reach.Init(world)

	actorP := world.p[actorID]

	// find nearest empty cells adjacent to enemies
	reach.Update(world, actorP)
	reachableIDs := make([]int, 0, len(enemySet))
	reachableD := -1
	for enemyID := range enemySet {
		ep := world.p[enemyID]
		reachableIDs, reachableD = world.updateAdjacentReach(reach, ep, reachableIDs, reachableD)
	}
	if len(reachableIDs) == 0 {
		if *verbose {
			log.Printf("nothing reachable for %s@%v#%v", string(world.r[actorID]), world.p[actorID], actorID)
		}
		return false
	}

	// choose the first, in reading order, most reachable enemy
	world.sortIDs(reachableIDs)
	targetID := reachableIDs[0]
	targetP := world.p[targetID]
	if *verbose {
		log.Printf(
			"target %s@%v#%v => %s@%v#%v",
			string(world.r[actorID]), world.p[actorID], actorID,
			string(world.r[targetID]), world.p[targetID], targetID,
		)
	}

	// move to the first, in reading order, nearest adjacent cell
	reach.Update(world, targetP)
	// world.placeReachLabels(reach, -1) XXX for debugging reachability scores
	reachableIDs, reachableD = world.updateAdjacentReach(reach, actorP, reachableIDs[:0], -1)
	if len(reachableIDs) == 0 {
		if *verbose {
			log.Printf("no nearest reachable for %s@%v#%v targeting %s@%v#%v",
				string(world.r[actorID]), actorP, actorID,
				string(world.r[targetID]), targetP, targetID,
			)
		}
		return false
	}
	world.sortIDs(reachableIDs)

	cellID := reachableIDs[0]
	if cellID == 0 {
		if *verbose {
			log.Printf("zero cell id! in %v", reachableIDs) // XXX inconceivable
		}
		return false
	}

	cellP := world.p[cellID]
	if *verbose {
		log.Printf(
			"move %s@%v#%v => %s@%v#%v",
			string(world.r[actorID]), world.p[actorID], actorID,
			string(world.r[cellID]), cellP, cellID,
		)
	}
	world.p[actorID] = cellP
	world.Index.Update(actorID, cellP)

	return true
}

func (world *gameWorld) enemyAt(p image.Point, enemySet map[int]struct{}) int {
	cur := world.Index.At(p)
	for cur.Next() {
		id := cur.I()
		if _, isEnemy := enemySet[id]; isEnemy {
			return id
		}
	}
	return 0
}

func (world *gameWorld) empty(p image.Point) (id int) {
	cur := world.Index.At(p)
	for cur.Next() {
		if world.t[cur.I()]&gameCollide != 0 {
			return 0
		}
		id = cur.I()
	}
	return id
}

func (world *gameWorld) placeReachLabels(reach reachabilityScore, z int) {
	ids := make([]int, 0, len(world.t))
	for id, t := range world.t {
		if t&gameLabel != 0 && world.z[id] == z {
			ids = append(ids, id)
		}
	}

	maxsc := world.bounds.Dx() + world.bounds.Dy()
	for p := reach.Min; p.Y < reach.Max.Y; p.Y++ {
		for p.X = reach.Min.X; p.X < reach.Max.X; p.X++ {
			i, _ := reach.Index(p)
			sc := reach.sc[i]
			if sc < 0 {
				continue
			}
			var id int
			if len(ids) > 0 {
				id = ids[len(ids)-1]
				ids = ids[:len(ids)-1]
			} else {
				id = world.createEntity(gameRender | gameLabel)
			}

			c := uint8(32 + (256-32)*sc/maxsc)
			world.p[id] = p
			world.z[id] = z
			world.r[id] = '0' + byte(sc%10)
			world.a[id] = ansi.RGB(c, c, 32).FG()
			world.Index.Update(id, p)
		}
	}

	for _, id := range ids {
		world.destroyEntity(id)
	}
}

type reachabilityScore struct {
	geom.RCore
	sc []int
}

func (r reachabilityScore) Get(p image.Point) int {
	if i, ok := r.Index(p); ok {
		return r.sc[i]
	}
	return -1
}

type searchCore struct {
	p []image.Point // frontier pos
	d []int         // frontier distance
	// u map[int]struct{} // frontier pruning

	sort sort.Interface
}

func (srch *searchCore) init(n int) {
	if cap(srch.p) < n {
		srch.p = make([]image.Point, 0, n)
	}
	if cap(srch.d) < n {
		srch.d = make([]int, 0, n)
	}
	// if len(srch.u) == 0 {
	// 	srch.u = make(map[int]struct{}, n)
	// } else {
	// 	for id := range srch.u {
	// 		delete(srch.u, id)
	// 	}
	// }
}

func (srch *searchCore) Len() int { return len(srch.p) }

func (srch *searchCore) push(p image.Point, d, id int) {
	srch.p = append(srch.p, p)
	srch.d = append(srch.d, d)
	// if id != 0 {
	// 	srch.u[id] = struct{}{}
	// }
	if srch.sort != nil && len(srch.p) > 1 {
		sort.Sort(srch.sort)
	}
}

func (srch *searchCore) pop() (image.Point, int) {
	i := len(srch.p) - 1
	pos, d := srch.p[i], srch.d[i]
	srch.p, srch.d = srch.p[:i], srch.d[:i]
	return pos, d
}

func (r *reachabilityScore) Init(world *gameWorld) {
	r.Rectangle = world.bounds
	r.Origin = world.bounds.Min
	r.Stride = world.bounds.Dx()
	r.sc = make([]int, world.bounds.Dy()*r.Stride)
}

func (r *reachabilityScore) Update(world *gameWorld, p image.Point) {
	for i := range r.sc {
		r.sc[i] = -1
	}

	// TODO factor out reachabilitySearch

	var srch searchCore // TODO re-use
	srch.init(len(world.p))
	srch.push(p, 0, world.empty(p))
	ri, _ := r.Index(p)
	r.sc[ri] = 0

	// sort reverse by distance, so we expand (a) closest point first
	srch.sort = sortPointRevBy{sortPointBy{&srch.p, &srch.d}}

	for srch.Len() > 0 {
		pos, d := srch.pop()

		// skip if it's not better than a prior score
		ri, _ := r.Index(pos)
		if sc := r.sc[ri]; sc >= 0 && d > sc {
			continue
		}

		// ... expand to any adjacent empty cells that are better than prior
		pts, cellIDs := world.adjacentCells(pos)
		for i, cellID := range cellIDs {
			if cellID != 0 {
				qp := pts[i]
				ri, _ := r.Index(qp)
				if sc := r.sc[ri]; sc < 0 || sc > d+1 {
					srch.push(qp, d+1, cellID)
					r.sc[ri] = d + 1
				}
			}
		}
	}
}

func (world *gameWorld) updateAdjacentReach(
	reach reachabilityScore,
	from image.Point,
	ids []int, best int,
) ([]int, int) {
	pts, cellIDs := world.adjacentCells(from)
	for i, cellID := range cellIDs {
		if cellID != 0 {
			d := reach.Get(pts[i])
			if d < 0 {
				continue
			}
			if best >= 0 && best < d {
				continue
			}
			if best > d {
				ids = ids[:0]
			}
			ids = append(ids, cellID)
			best = d
		}
	}
	return ids, best
}

func (world *gameWorld) adjacentCells(p image.Point) (pts [4]image.Point, cellIDs [4]int) {
	pts[0] = p.Add(image.Pt(0, -1))
	pts[1] = p.Add(image.Pt(-1, 0))
	pts[2] = p.Add(image.Pt(1, 0))
	pts[3] = p.Add(image.Pt(0, 1))
	cellIDs[0] = world.empty(pts[0])
	cellIDs[1] = world.empty(pts[1])
	cellIDs[2] = world.empty(pts[2])
	cellIDs[3] = world.empty(pts[3])
	return pts, cellIDs
}

func manhattanDistance(a, b image.Point) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (world *gameWorld) sortIDs(ids []int) {
	sort.Slice(ids, func(i, j int) bool {
		idi, idj := ids[i], ids[j]
		pi, pj := world.p[idi], world.p[idj]
		if pi.Y < pj.Y {
			return true
		}
		return pi.Y == pj.Y && pi.X < pj.X
	})
}

func (world *gameWorld) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	return false, nil
}

type sortBy struct{ a, b []int }

func (s sortBy) Len() int               { return len(s.a) }
func (s sortBy) Less(i int, j int) bool { return s.b[i] < s.b[j] }
func (s sortBy) Swap(i int, j int) {
	s.a[i], s.a[j] = s.a[j], s.a[i]
	s.b[i], s.b[j] = s.b[j], s.b[i]
}

type sortPointBy struct {
	a *[]image.Point
	b *[]int
}

func (s sortPointBy) Len() int               { return len(*(s.a)) }
func (s sortPointBy) Less(i int, j int) bool { return (*s.b)[i] < (*s.b)[j] }
func (s sortPointBy) Swap(i int, j int) {
	(*s.a)[i], (*s.a)[j] = (*s.a)[j], (*s.a)[i]
	(*s.b)[i], (*s.b)[j] = (*s.b)[j], (*s.b)[i]
}

type sortPointRevBy struct{ sortPointBy }

func (s sortPointRevBy) Less(i int, j int) bool { return (*s.b)[i] > (*s.b)[j] }
